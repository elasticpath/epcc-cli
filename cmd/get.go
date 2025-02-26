package cmd

import (
	"context"
	gojson "encoding/json"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/elasticpath/epcc-cli/external/rest"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"time"
)

const singularResourceRequest = 0
const collectionResourceRequest = 1

func NewGetCommand(parentCmd *cobra.Command) func() {
	overrides := &httpclient.HttpParameterOverrides{
		QueryParameters: nil,
		OverrideUrlPath: "",
	}

	// Ensure that any new options here are added to the resetFunc
	var outputJq = ""
	var compactOutput = false
	var noBodyPrint = false
	var retryWhileJQ = ""
	var retryWhileJQMaxAttempts = uint16(1200)
	var ifAliasExists = ""
	var ifAliasDoesNotExist = ""
	var skipAliases = false
	var repeat uint32 = 1
	var repeatDelay uint32 = 100
	var ignoreErrors = false

	resetFunc := func() {
		overrides.QueryParameters = nil
		overrides.OverrideUrlPath = ""
		outputJq = ""
		compactOutput = false
		noBodyPrint = false
		retryWhileJQ = ""
		retryWhileJQMaxAttempts = uint16(1200)
		ifAliasExists = ""
		ifAliasDoesNotExist = ""
		skipAliases = false
		repeat = 1
		repeatDelay = 100
		ignoreErrors = false

	}

	var getCmd = &cobra.Command{
		Use:          "get",
		Short:        "Retrieves either a single or all resources",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("please specify a resource, epcc get [RESOURCE], see epcc get --help")
			} else {
				return fmt.Errorf("invalid resource [%s] specified, see all with epcc get --help", args[0])
			}
		},
	}

	for _, resource := range resources.GetPluralResources() {
		resource := resource

		for i := 0; i < 2; i++ {
			i := i
			//usageString := ""
			resourceName := ""
			completionVerb := 0
			usageGetType := ""
			var urlInfo *resources.CrudEntityInfo = nil

			switch i {
			case singularResourceRequest:
				if resource.GetCollectionInfo == nil {
					continue
				}

				resourceName = resource.PluralName
				//usageString = resource.PluralName

				urlInfo = resource.GetCollectionInfo
				completionVerb = completion.GetAll
				usageGetType = "all (or a single page) of"

			case collectionResourceRequest:
				if resource.GetEntityInfo == nil {
					continue
				}
				//usageString = resource.SingularName
				resourceName = resource.SingularName

				urlInfo = resource.GetEntityInfo
				completionVerb = completion.Get
				usageGetType = "a single"
			}

			resourceUrl := urlInfo.Url

			newCmd := &cobra.Command{
				Use: GetGetUsageString(resourceName, resourceUrl, completionVerb, resource),
				// The replace all is a hack for the moment the URL could be made nicer
				Short:   GetGetShort(resourceUrl),
				Long:    GetGetLong(resourceName, resourceUrl, usageGetType, completionVerb, urlInfo, resource),
				Example: GetGetExample(resourceName, resourceUrl, usageGetType, completionVerb, urlInfo, resource),
				Args:    GetArgFunctionForUrl(resourceName, resourceUrl),
				RunE: func(cmd *cobra.Command, args []string) error {

					c := func(cmd *cobra.Command, args []string) error {

						if ifAliasExists != "" {
							aliasId := aliases.ResolveAliasValuesOrReturnIdentity(resource.JsonApiType, resource.AlternateJsonApiTypesForAliases, ifAliasExists, "id")

							if aliasId == ifAliasExists {
								// If the aliasId is the same as requested, it means an alias did not exist.
								log.Infof("Alias [%s] does not exist, not continuing run", ifAliasExists)
								return nil
							}
						}

						if ifAliasDoesNotExist != "" {
							aliasId := aliases.ResolveAliasValuesOrReturnIdentity(resource.JsonApiType, resource.AlternateJsonApiTypesForAliases, ifAliasDoesNotExist, "id")

							if aliasId != ifAliasDoesNotExist {
								// If the aliasId is different than the request then it does exist.
								log.Infof("Alias [%s] does exist (value: %s), not continuing run", ifAliasDoesNotExist, aliasId)
								return nil
							}
						}

						var body string
						var err error
						if retryWhileJQMaxAttempts == 0 {
							return fmt.Errorf("--retry-while-jq-max-attempts must be greater than 0")
						}

						retriesFailedError := fmt.Errorf("Maximum number of retries hit %d and condition [%s] always true", retryWhileJQMaxAttempts, retryWhileJQ)

						for attempt := uint16(0); attempt < retryWhileJQMaxAttempts; attempt++ {
							body, err = rest.GetInternal(context.Background(), overrides, append([]string{resourceName}, args...), skipAliases)
							if retryWhileJQ == "" {
								retriesFailedError = nil
								break
							}

							resultOfRetryWhileJQ, err := json.RunJQOnStringWithArray(retryWhileJQ, body)

							if err != nil {
								break
							}

							if len(resultOfRetryWhileJQ) > 0 {
								if result, ok := resultOfRetryWhileJQ[0].(bool); ok {
									if result {
										time.Sleep(3 * time.Second)
										continue
									}
								}
							}

							log.Infof("Result of JQ [%s] was: %v, retries complete", retryWhileJQ, resultOfRetryWhileJQ)
							retriesFailedError = nil
							break

						}

						if err != nil {
							return err
						}

						if outputJq != "" {
							output, err := json.RunJQOnStringWithArray(outputJq, body)

							if err != nil {
								return err
							}
							for _, outputLine := range output {
								outputJson, err := gojson.Marshal(outputLine)

								if err != nil {
									return err
								}

								err = json.PrintJsonToStdout(string(outputJson))

								if err != nil {
									return err
								}
							}

							return retriesFailedError
						}

						if noBodyPrint {
							return retriesFailedError
						} else {
							if compactOutput {
								body, err = json.Compact(body)

								if err != nil {
									return err
								}
							}

							printError := json.PrintJsonToStdout(body)

							if retriesFailedError != nil {
								return retriesFailedError
							}

							return printError
						}
					}

					return repeater(c, repeat, repeatDelay, cmd, args, ignoreErrors)

				},
				ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

					if resourceUrl == "" {
						return []string{}, cobra.ShellCompDirectiveNoFileComp
					}

					idCount, err := resources.GetNumberOfVariablesNeeded(resourceUrl)

					if err != nil {
						return []string{}, cobra.ShellCompDirectiveNoFileComp
					}

					if len(args) < idCount {
						// Must be for a resource completion
						types, err := resources.GetTypesOfVariablesNeeded(resourceUrl)

						if err != nil {
							return []string{}, cobra.ShellCompDirectiveNoFileComp
						}

						typeIdxNeeded := len(args)

						if completionResource, ok := resources.GetResourceByName(types[typeIdxNeeded]); ok {
							return completion.Complete(completion.Request{
								Type:     completion.CompleteAlias,
								Resource: completionResource,
							})
						}

					} else if len(args) >= idCount { // Arg is after IDs
						if (len(args)-idCount)%2 == 0 { // This is a query param key
							return completion.Complete(completion.Request{
								Type:     completion.CompleteQueryParamKey,
								Resource: resource,
								Verb:     completionVerb,
							})

						} else {
							return completion.Complete(completion.Request{
								Type:       completion.CompleteQueryParamValue,
								Resource:   resource,
								Verb:       completionVerb,
								QueryParam: args[len(args)-1],
								ToComplete: toComplete,
							})

						}
					}

					return []string{}, cobra.ShellCompDirectiveNoFileComp
				},
			}

			getCmd.AddCommand(newCmd)
		}
	}

	getCmd.PersistentFlags().BoolVarP(&noBodyPrint, "silent", "s", false, "Don't print the body on success")
	getCmd.PersistentFlags().StringVar(&overrides.OverrideUrlPath, "override-url-path", "", "Override the URL that will be used for the Request")
	getCmd.PersistentFlags().StringSliceVarP(&overrides.QueryParameters, "query-parameters", "q", []string{}, "Pass in key=value an they will be added as query parameters")
	getCmd.PersistentFlags().StringVarP(&outputJq, "output-jq", "", "", "A jq expression, if set we will restrict output to only this")
	getCmd.PersistentFlags().BoolVarP(&compactOutput, "compact", "", false, "Hides some of the boiler plate keys and empty fields, etc...")
	getCmd.PersistentFlags().BoolVarP(&ignoreErrors, "ignore-errors", "", false, "Don't return non zero on an error")
	getCmd.PersistentFlags().StringVarP(&retryWhileJQ, "retry-while-jq", "", "", "A jq expression, if set and returns true we will retry the get command (see manual for examples)")
	getCmd.PersistentFlags().Uint16VarP(&retryWhileJQMaxAttempts, "retry-while-jq-max-attempts", "", 1200, "The maximum number of attempts we will retry with jq")
	getCmd.PersistentFlags().StringVarP(&ifAliasExists, "if-alias-exists", "", "", "If the alias exists we will run this command, otherwise exit with no error")
	getCmd.PersistentFlags().StringVarP(&ifAliasDoesNotExist, "if-alias-does-not-exist", "", "", "If the alias does not exist we will run this command, otherwise exit with no error")
	getCmd.PersistentFlags().BoolVarP(&skipAliases, "skip-alias-processing", "", false, "if set, we don't process the response for aliases")
	getCmd.MarkFlagsMutuallyExclusive("if-alias-exists", "if-alias-does-not-exist")
	getCmd.PersistentFlags().Uint32VarP(&repeat, "repeat", "", 1, "Number of times to repeat the command")
	getCmd.PersistentFlags().Uint32VarP(&repeatDelay, "repeat-delay", "", 100, "Delay (in ms) between repeats")

	_ = getCmd.RegisterFlagCompletionFunc("output-jq", jqCompletionFunc)

	parentCmd.AddCommand(getCmd)

	return resetFunc
}
