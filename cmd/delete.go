package cmd

import (
	"context"
	"fmt"
	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/elasticpath/epcc-cli/external/rest"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewDeleteCommand(parentCmd *cobra.Command) func() {

	var deleteCmd = &cobra.Command{
		Use:          "delete",
		Short:        "Deletes a resource",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("please specify a resource, epcc delete [RESOURCE], see epcc delete --help")
			} else {
				return fmt.Errorf("invalid resource [%s] specified, see all with epcc delete --help", args[0])
			}
		},
	}

	overrides := &httpclient.HttpParameterOverrides{
		QueryParameters: nil,
		OverrideUrlPath: "",
	}

	// Ensure that any new options here are added to the resetFunc
	var allow404 = false
	var ifAliasExists = ""
	var ifAliasDoesNotExist = ""
	var repeat uint32 = 1
	var repeatDelay uint32 = 100
	var ignoreErrors = false
	var noBodyPrint = false
	var logOnSuccess = ""
	var logOnFailure = ""

	resetFunc := func() {
		overrides.QueryParameters = nil
		overrides.OverrideUrlPath = ""
		allow404 = false
		ifAliasExists = ""
		ifAliasDoesNotExist = ""
		noBodyPrint = false
		repeat = 1
		repeatDelay = 100
		ignoreErrors = false
		logOnSuccess = ""
		logOnFailure = ""
	}

	e := config.GetEnv()
	hiddenResources := map[string]struct{}{}
	for _, v := range e.EPCC_CLI_DISABLE_RESOURCES {
		hiddenResources[v] = struct{}{}
	}

	for _, resource := range resources.GetPluralResources() {
		if resource.DeleteEntityInfo == nil {
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

		var deleteResourceCommand = &cobra.Command{
			Use:     GetDeleteUsage(resource),
			Short:   GetDeleteShort(resource),
			Long:    GetDeleteLong(resource),
			Example: GetDeleteExample(resource),
			Args:    GetArgFunctionForDelete(resource),
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

					body, err := rest.DeleteInternal(context.Background(), overrides, allow404, append([]string{resourceName}, args...))

					if err != nil {
						if body != "" {
							if !noBodyPrint {
								json.PrintJson(body)
							}
						}
						return err
					}

					if noBodyPrint {
						return nil
					} else {
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

				// len(args) == 1 means first id
				// lens(args) == 2 means second id.

				// Replace ids with args in resourceURL
				if resource.DeleteEntityInfo == nil {
					return []string{}, cobra.ShellCompDirectiveNoFileComp
				}

				idCount, err := resources.GetNumberOfVariablesNeeded(resource.DeleteEntityInfo.Url)

				if err != nil {
					return []string{}, cobra.ShellCompDirectiveNoFileComp
				}

				if len(args) < idCount {
					// Must be for a resource completion
					types, err := resources.GetTypesOfVariablesNeeded(resource.DeleteEntityInfo.Url)

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
				} else {
					if (len(args)-idCount)%2 == 0 { // This is an attribute key
						usedAttributes := make(map[string]int)
						for i := idCount; i < len(args); i = i + 2 {
							usedAttributes[args[i]] = 0
						}
						return completion.Complete(completion.Request{
							Type:       completion.CompleteAttributeKey,
							Resource:   resource,
							Attributes: usedAttributes,
							Verb:       completion.Delete,
						})
					} else { // This is an attribute value
						return completion.Complete(completion.Request{
							Type:           completion.CompleteAttributeValue,
							Resource:       resource,
							Verb:           completion.Delete,
							Attribute:      args[len(args)-1],
							ToComplete:     toComplete,
							AllowTemplates: true,
						})
					}
				}

				return []string{}, cobra.ShellCompDirectiveNoFileComp
			},
		}

		deleteCmd.AddCommand(deleteResourceCommand)
	}

	deleteCmd.PersistentFlags().StringVar(&overrides.OverrideUrlPath, "override-url-path", "", "Override the URL that will be used for the Request")
	deleteCmd.PersistentFlags().StringSliceVarP(&overrides.QueryParameters, "query-parameters", "q", []string{}, "Pass in key=value an they will be added as query parameters")
	deleteCmd.PersistentFlags().BoolVar(&allow404, "allow-404", allow404, "If set 404's will not be treated as errors")
	deleteCmd.PersistentFlags().StringVarP(&ifAliasExists, "if-alias-exists", "", "", "If the alias exists we will run this command, otherwise exit with no error")
	deleteCmd.PersistentFlags().StringVarP(&ifAliasDoesNotExist, "if-alias-does-not-exist", "", "", "If the alias does not exist we will run this command, otherwise exit with no error")
	deleteCmd.PersistentFlags().BoolVarP(&noBodyPrint, "silent", "s", false, "Don't print the body on success")
	deleteCmd.MarkFlagsMutuallyExclusive("if-alias-exists", "if-alias-does-not-exist")
	deleteCmd.PersistentFlags().Uint32VarP(&repeat, "repeat", "", 1, "Number of times to repeat the command")
	deleteCmd.PersistentFlags().Uint32VarP(&repeatDelay, "repeat-delay", "", 100, "Delay (in ms) between repeats")
	deleteCmd.PersistentFlags().BoolVarP(&ignoreErrors, "ignore-errors", "", false, "Don't return non zero on an error")
	deleteCmd.PersistentFlags().StringVarP(&logOnSuccess, "log-on-success", "", "", "Output the following message as an info if the result is successful")
	deleteCmd.PersistentFlags().StringVarP(&logOnFailure, "log-on-failure", "", "", "Output the following message as an error if the result fails")
	parentCmd.AddCommand(deleteCmd)

	return resetFunc
}
