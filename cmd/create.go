package cmd

import (
	gojson "encoding/json"
	"fmt"
	"strings"

	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/clictx"

	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/elasticpath/epcc-cli/external/rest"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewCreateCommand(parentCmd *cobra.Command) func() {

	var createCmd = &cobra.Command{
		Use:   "create",
		Short: "Creates a resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("please specify a resource, epcc create [RESOURCE], see epcc create --help")
			} else {
				return fmt.Errorf("invalid resource [%s] specified, see all with epcc create --help", args[0])
			}
		},
	}

	overrides := &httpclient.HttpParameterOverrides{
		QueryParameters: nil,
		OverrideUrlPath: "",
	}

	// Ensure that any new options here are added to the resetFunc
	var autoFillOnCreate = false
	var noBodyPrint = false
	var outputKeyValue = false
	var outputJq = ""
	var compactOutput = true
	var setAlias = ""
	var ifAliasExists = ""
	var ifAliasDoesNotExist = ""
	var skipAliases = false
	var repeat uint32 = 1
	var repeatDelay uint32 = 100
	var ignoreErrors = false
	var logOnSuccess = ""
	var logOnFailure = ""
	var disableConstants = false
	var data = ""

	resetFunc := func() {
		autoFillOnCreate = false
		noBodyPrint = false
		outputKeyValue = false
		outputJq = ""
		setAlias = ""
		ifAliasExists = ""
		ifAliasDoesNotExist = ""
		overrides.OverrideUrlPath = ""
		overrides.QueryParameters = nil
		skipAliases = false
		compactOutput = false
		repeat = 1
		repeatDelay = 100
		ignoreErrors = false
		logOnSuccess = ""
		logOnFailure = ""
		disableConstants = false
		data = ""
	}

	e := config.GetEnv()
	hiddenResources := map[string]struct{}{}
	for _, v := range e.EPCC_CLI_DISABLE_RESOURCES {
		hiddenResources[v] = struct{}{}
	}

	for _, resource := range resources.GetPluralResources() {

		if resource.CreateEntityInfo == nil {
			continue
		}

		if _, ok := hiddenResources[resource.SingularName]; ok {
			log.Tracef("Hiding resource %s", resource.SingularName)
			continue
		}

		if _, ok := hiddenResources[resource.PluralName]; ok {
			log.Tracef("Hiding resource %s", resource.SingularName)
			continue
		}

		resource := resource
		resourceName := resource.SingularName

		var createResourceCmd = &cobra.Command{
			Use:     GetCreateUsageString(resource),
			Short:   GetCreateShort(resource),
			Long:    GetCreateLong(resource),
			Example: GetCreateExample(resource),
			Args:    GetArgFunctionForCreate(resource),
			RunE: func(cmd *cobra.Command, args []string) error {
				c := func(cmd *cobra.Command, args []string) error {
					if ifAliasExists != "" {
						aliasId := aliases.ResolveAliasValuesOrReturnIdentity(resource.JsonApiType, resource.AlternateJsonApiTypesForAliases, ifAliasExists, "id")

						if aliasId == ifAliasExists {
							// If the aliasId is the same as requested, it means an alias did not exist.
							log.Debugf("Alias [%s] does not exist, not continuing run", ifAliasExists)
							return nil
						}
					}

					if ifAliasDoesNotExist != "" {
						aliasId := aliases.ResolveAliasValuesOrReturnIdentity(resource.JsonApiType, resource.AlternateJsonApiTypesForAliases, ifAliasDoesNotExist, "id")

						if aliasId != ifAliasDoesNotExist {
							// If the aliasId is different than the request then it does exist.
							log.Debugf("Alias [%s] does exist (value: %s), not continuing run", ifAliasDoesNotExist, aliasId)
							return nil
						}
					}

					body, err := rest.CreateInternal(clictx.Ctx, overrides, append([]string{resourceName}, args...), autoFillOnCreate, setAlias, skipAliases, disableConstants, data)

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

						return nil
					}

					if noBodyPrint {
						return nil
					} else if outputKeyValue {
						return json.PrintJsonAsKeyValue(body)
					} else {
						if compactOutput {
							body, err = json.Compact(body)

							if err != nil {
								return err
							}
						}

						return json.PrintJson(body)
					}
				}

				res := repeater(c, repeat, repeatDelay, cmd, args, ignoreErrors)

				if res != nil {
					if logOnFailure != "" {
						log.Errorf("%s", logOnFailure)
					}
				} else {
					if logOnSuccess != "" {
						log.Infof("%s", logOnSuccess)
					}
				}
				return res
			},

			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				// Find Resource
				resource, ok := resources.GetResourceByName(resourceName)
				if ok {
					if resource.CreateEntityInfo != nil {
						resourceURL := resource.CreateEntityInfo.Url
						idCount, _ := resources.GetNumberOfVariablesNeeded(resourceURL)
						if len(args)-idCount >= 0 { // Arg is after IDs
							if (len(args)-idCount)%2 == 0 { // This is an attribute key
								usedAttributes := make(map[string]string)
								for i := idCount; i < len(args); i = i + 2 {
									if i+1 <= len(args) {
										usedAttributes[args[i]] = args[i+1]
									} else {
										usedAttributes[args[i]] = ""
									}
								}

								// I think this allows you to complete the current argument
								// This is necessary because if you are using something with a wildcard or regex
								// You won't see it in the attribute list, and therefore it won't be able to auto complete it.
								// I now think this does nothing.
								toComplete := strings.ReplaceAll(toComplete, "<ENTER>", "")
								if toComplete != "" {
									usedAttributes[toComplete] = ""
								}

								return completion.Complete(completion.Request{
									Type:       completion.CompleteAttributeKey,
									Resource:   resource,
									Attributes: usedAttributes,
									Verb:       completion.Create,
									ToComplete: toComplete,
								})
							} else { // This is an attribute value
								return completion.Complete(completion.Request{
									Type:           completion.CompleteAttributeValue,
									Resource:       resource,
									Verb:           completion.Create,
									Attribute:      args[len(args)-1],
									ToComplete:     toComplete,
									AllowTemplates: true,
								})
							}
						} else {
							// Arg is in IDS
							// Must be for a resource completion
							types, err := resources.GetTypesOfVariablesNeeded(resourceURL)

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
						}
					}
				}

				return []string{}, cobra.ShellCompDirectiveNoFileComp
			},
		}

		createCmd.AddCommand(createResourceCmd)
	}

	parentCmd.AddCommand(createCmd)
	createCmd.PersistentFlags().StringVar(&overrides.OverrideUrlPath, "override-url-path", "", "Override the URL that will be used for the Request")
	createCmd.PersistentFlags().BoolVarP(&autoFillOnCreate, "auto-fill", "", false, "Auto generate value for fields")
	createCmd.PersistentFlags().BoolVarP(&noBodyPrint, "silent", "s", false, "Don't print the body on success")
	createCmd.PersistentFlags().BoolVarP(&outputKeyValue, "output-key-val", "", false, "Outputs the result in epcc-cli json key/value format")
	createCmd.PersistentFlags().StringSliceVarP(&overrides.QueryParameters, "query-parameters", "q", []string{}, "Pass in key=value an they will be added as query parameters")
	createCmd.PersistentFlags().StringVarP(&outputJq, "output-jq", "", "", "A jq expression, if set we will restrict output to only this")
	createCmd.PersistentFlags().BoolVarP(&compactOutput, "compact", "", false, "Hides some of the boiler plate keys and empty fields, etc...")
	createCmd.PersistentFlags().BoolVarP(&ignoreErrors, "ignore-errors", "", false, "Don't return non-zero on an error")
	createCmd.PersistentFlags().StringVarP(&setAlias, "save-as-alias", "", "", "A name to save the created resource as")
	createCmd.PersistentFlags().StringVarP(&ifAliasExists, "if-alias-exists", "", "", "If the alias exists we will run this command, otherwise exit with no error")
	createCmd.PersistentFlags().StringVarP(&ifAliasDoesNotExist, "if-alias-does-not-exist", "", "", "If the alias does not exist we will run this command, otherwise exit with no error")
	createCmd.PersistentFlags().BoolVarP(&skipAliases, "skip-alias-processing", "", false, "if set, we don't process the response for aliases")
	createCmd.MarkFlagsMutuallyExclusive("if-alias-exists", "if-alias-does-not-exist")
	createCmd.PersistentFlags().Uint32VarP(&repeat, "repeat", "", 1, "Number of times to repeat the command")
	createCmd.PersistentFlags().Uint32VarP(&repeatDelay, "repeat-delay", "", 100, "Delay (in ms) between repeats")

	createCmd.PersistentFlags().BoolVarP(&disableConstants, "no-auto-constants", "", false, "Disable setting of known constant values in the request body")
	createCmd.PersistentFlags().StringVarP(&logOnSuccess, "log-on-success", "", "", "Output the following message as an info if the result is successful")
	createCmd.PersistentFlags().StringVarP(&logOnFailure, "log-on-failure", "", "", "Output the following message as an error if the result fails")
	createCmd.PersistentFlags().StringVarP(&data, "data", "d", "", "Raw JSON data to use as the request body. If provided, positional arguments will be ignored.")

	createCmd.MarkFlagsMutuallyExclusive("output-key-val", "output-jq", "silent", "compact")
	_ = createCmd.RegisterFlagCompletionFunc("output-jq", jqCompletionFunc)

	return resetFunc
}
