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
	"github.com/elasticpath/epcc-cli/external/shutdown"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"net/url"
	"strings"
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
	var noBodyPrint = false
	var retryWhileJQ = ""
	var retryWhileJQMaxAttempts = uint16(1200)
	var ifAliasExists = ""
	var ifAliasDoesNotExist = ""
	var skipAliases = false

	resetFunc := func() {
		overrides.QueryParameters = nil
		overrides.OverrideUrlPath = ""
		outputJq = ""
		noBodyPrint = false
		retryWhileJQ = ""
		retryWhileJQMaxAttempts = uint16(1200)
		ifAliasExists = ""
		ifAliasDoesNotExist = ""
		skipAliases = false
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
						body, err = getInternal(context.Background(), overrides, append([]string{resourceName}, args...), skipAliases)
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

							err = json.PrintJson(string(outputJson))

							if err != nil {
								return err
							}
						}

						return retriesFailedError
					}

					if noBodyPrint {
						return retriesFailedError
					} else {
						printError := json.PrintJson(body)

						if retriesFailedError != nil {
							return retriesFailedError
						}

						return printError
					}

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
	getCmd.PersistentFlags().StringVarP(&retryWhileJQ, "retry-while-jq", "", "", "A jq expression, if set and returns true we will retry the get command (see manual for examples)")
	getCmd.PersistentFlags().Uint16VarP(&retryWhileJQMaxAttempts, "retry-while-jq-max-attempts", "", 1200, "The maximum number of attempts we will retry with jq")
	getCmd.PersistentFlags().StringVarP(&ifAliasExists, "if-alias-exists", "", "", "If the alias exists we will run this command, otherwise exit with no error")
	getCmd.PersistentFlags().StringVarP(&ifAliasDoesNotExist, "if-alias-does-not-exist", "", "", "If the alias does not exist we will run this command, otherwise exit with no error")
	getCmd.PersistentFlags().BoolVarP(&skipAliases, "skip-alias-processing", "", false, "if set, we don't process the response for aliases")
	getCmd.MarkFlagsMutuallyExclusive("if-alias-exists", "if-alias-does-not-exist")
	_ = getCmd.RegisterFlagCompletionFunc("output-jq", jqCompletionFunc)

	parentCmd.AddCommand(getCmd)

	return resetFunc
}

func getInternal(ctx context.Context, overrides *httpclient.HttpParameterOverrides, args []string, skipAliases bool) (string, error) {
	resp, err := getResource(ctx, overrides, args)

	if err != nil {
		return "", err
	} else if resp == nil {
		return "", fmt.Errorf("got nil response")
	}

	if resp.Body != nil {
		defer resp.Body.Close()

		// Print the body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		// Check if error response
		if resp.StatusCode >= 400 && resp.StatusCode <= 600 {
			json.PrintJson(string(body))
			return "", fmt.Errorf(resp.Status)
		}

		if !skipAliases {
			aliases.SaveAliasesForResources(string(body))
		}

		return string(body), nil
	} else {
		return "", nil
	}
}

func getUrl(resource resources.Resource, args []string) (*resources.CrudEntityInfo, error) {

	if resource.GetCollectionInfo == nil && resource.GetEntityInfo == nil {
		return nil, fmt.Errorf("resource %s doesn't support GET", args[0])
	} else if resource.GetCollectionInfo != nil && resource.GetEntityInfo == nil {
		return resource.GetCollectionInfo, nil
	} else if resource.GetCollectionInfo == nil && resource.GetEntityInfo != nil {
		return resource.GetEntityInfo, nil
	} else {
		if _, ok := resources.GetPluralResources()[args[0]]; ok {
			return resource.GetCollectionInfo, nil
		} else {
			return resource.GetEntityInfo, nil
		}
	}
}

func getResource(ctx context.Context, overrides *httpclient.HttpParameterOverrides, args []string) (*http.Response, error) {
	shutdown.OutstandingOpCounter.Add(1)
	defer shutdown.OutstandingOpCounter.Done()

	// Find Resource
	resource, ok := resources.GetResourceByName(args[0])
	if !ok {
		return nil, fmt.Errorf("could not find resource %s", args[0])
	}

	var idCount int

	resourceUrlInfo, err2 := getUrl(resource, args)
	if err2 != nil {
		return nil, err2
	}

	idCount, err := resources.GetNumberOfVariablesNeeded(resourceUrlInfo.Url)

	if err != nil {
		return nil, err
	}

	// Replace ids with args in resourceURL
	resourceURL, err := resources.GenerateUrl(resourceUrlInfo, args[1:], true)

	if err != nil {
		return nil, err
	}

	if overrides.OverrideUrlPath != "" {
		log.Warnf("Overriding URL Path from %s to %s", resourceURL, overrides.OverrideUrlPath)
		resourceURL = overrides.OverrideUrlPath
	}

	// Add remaining args as query params
	params := url.Values{}
	for i := idCount + 1; i+1 < len(args); i = i + 2 {
		params.Add(args[i], args[i+1])
	}

	if (idCount-len(args)+1)%2 != 0 {
		log.Warnf("Extra argument at the end of the command %s", args[len(args)-1])
	}

	for _, v := range overrides.QueryParameters {
		keyAndValue := strings.SplitN(v, "=", 2)
		if len(keyAndValue) != 2 {
			return nil, fmt.Errorf("Could not parse query parameter %v, all query parameters should be a key and value format", keyAndValue)
		}
		params.Add(keyAndValue[0], keyAndValue[1])
	}

	// Submit request
	resp, err := httpclient.DoRequest(ctx, "GET", resourceURL, params.Encode(), nil)

	if err != nil {
		return nil, fmt.Errorf("got error %s", err.Error())
	}

	return resp, nil
}
