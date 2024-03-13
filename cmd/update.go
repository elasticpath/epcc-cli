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
	"net/url"
	"strings"
)

func NewUpdateCommand(parentCmd *cobra.Command) func() {
	overrides := &httpclient.HttpParameterOverrides{
		QueryParameters: nil,
		OverrideUrlPath: "",
	}

	// Ensure that any new options here are added to the resetFunc
	var outputJq = ""
	var compactOutput = false
	var noBodyPrint = false
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
		ifAliasExists = ""
		ifAliasDoesNotExist = ""
		skipAliases = false
		repeat = 1
		repeatDelay = 100
		ignoreErrors = false
	}

	var updateCmd = &cobra.Command{
		Use:          "update",
		Short:        "Updates a resource",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("please specify a resource, epcc update [RESOURCE], see epcc update --help")
			} else {
				return fmt.Errorf("invalid resource [%s] specified, see all with epcc update --help", args[0])
			}
		},
	}
	for _, resource := range resources.GetPluralResources() {
		resource := resource
		resourceName := resource.SingularName
		if resource.UpdateEntityInfo == nil {
			continue
		}

		var updateResourceCmd = &cobra.Command{
			Use:     GetUpdateUsage(resource),
			Short:   GetUpdateShort(resource),
			Long:    GetUpdateLong(resource),
			Example: GetUpdateExample(resource),
			Args:    GetArgFunctionForUpdate(resource),
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

					body, err := updateInternal(context.Background(), overrides, skipAliases, append([]string{resourceName}, args...))

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

				return repeater(c, repeat, repeatDelay, cmd, args, ignoreErrors)
			},

			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				// Find Resource
				resourceURL := resource.UpdateEntityInfo.Url
				idCount, _ := resources.GetNumberOfVariablesNeeded(resourceURL)
				if len(args)-idCount >= 0 { // Arg is after IDs
					if (len(args)-idCount)%2 == 0 { // This is an attribute key
						usedAttributes := make(map[string]int)
						for i := idCount; i < len(args); i = i + 2 {
							usedAttributes[args[i]] = 0
						}
						return completion.Complete(completion.Request{
							Type:       completion.CompleteAttributeKey,
							Resource:   resource,
							Attributes: usedAttributes,
							Verb:       completion.Update,
						})
					} else { // This is an attribute value
						return completion.Complete(completion.Request{
							Type:           completion.CompleteAttributeValue,
							Resource:       resource,
							Verb:           completion.Update,
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
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			},
		}

		updateCmd.AddCommand(updateResourceCmd)
	}

	updateCmd.PersistentFlags().StringVar(&overrides.OverrideUrlPath, "override-url-path", "", "Override the URL that will be used for the Request")
	updateCmd.PersistentFlags().StringSliceVarP(&overrides.QueryParameters, "query-parameters", "q", []string{}, "Pass in key=value an they will be added as query parameters")
	updateCmd.PersistentFlags().BoolVarP(&noBodyPrint, "silent", "s", false, "Don't print the body on success")
	updateCmd.PersistentFlags().StringVarP(&outputJq, "output-jq", "", "", "A jq expression, if set we will restrict output to only this")
	updateCmd.PersistentFlags().BoolVarP(&compactOutput, "compact", "", false, "Hides some of the boiler plate keys and empty fields, etc...")
	updateCmd.PersistentFlags().BoolVarP(&ignoreErrors, "ignore-errors", "", false, "Don't return non zero on an error")
	updateCmd.PersistentFlags().StringVarP(&ifAliasExists, "if-alias-exists", "", "", "If the alias exists we will run this command, otherwise exit with no error")
	updateCmd.PersistentFlags().StringVarP(&ifAliasDoesNotExist, "if-alias-does-not-exist", "", "", "If the alias does not exist we will run this command, otherwise exit with no error")
	updateCmd.MarkFlagsMutuallyExclusive("if-alias-exists", "if-alias-does-not-exist")
	updateCmd.PersistentFlags().BoolVarP(&skipAliases, "skip-alias-processing", "", false, "if set, we don't process the response for aliases")
	_ = updateCmd.RegisterFlagCompletionFunc("output-jq", jqCompletionFunc)
	updateCmd.PersistentFlags().Uint32VarP(&repeat, "repeat", "", 1, "Number of times to repeat the command")
	updateCmd.PersistentFlags().Uint32VarP(&repeatDelay, "repeat-delay", "", 100, "Delay (in ms) between repeats")

	parentCmd.AddCommand(updateCmd)

	return resetFunc

}

func updateInternal(ctx context.Context, overrides *httpclient.HttpParameterOverrides, skipAliases bool, args []string) (string, error) {
	shutdown.OutstandingOpCounter.Add(1)
	defer shutdown.OutstandingOpCounter.Done()

	// Find Resource
	resource, ok := resources.GetResourceByName(args[0])
	if !ok {
		return "", fmt.Errorf("could not find resource %s", args[0])
	}

	if resource.UpdateEntityInfo == nil {
		return "", fmt.Errorf("resource %s doesn't support UPDATE", args[0])
	}

	// Count ids in UpdateEntity
	resourceUrlInfo := resource.UpdateEntityInfo
	idCount, err := resources.GetNumberOfVariablesNeeded(resourceUrlInfo.Url)
	if err != nil {
		return "", err
	}

	// Replace ids with args in resourceURL
	resourceURL, err := resources.GenerateUrl(resourceUrlInfo, args[1:], true)
	if err != nil {
		return "", err
	}

	if overrides.OverrideUrlPath != "" {
		log.Warnf("Overriding URL Path from %s to %s", resourceURL, overrides.OverrideUrlPath)
		resourceURL = overrides.OverrideUrlPath
	}

	args = append(args, "type", resource.JsonApiType)
	// Create the body from remaining args
	body, err := json.ToJson(args[(idCount+1):], resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes, true)
	if err != nil {
		return "", err
	}

	params := url.Values{}

	for _, v := range overrides.QueryParameters {
		keyAndValue := strings.SplitN(v, "=", 2)
		if len(keyAndValue) != 2 {
			return "", fmt.Errorf("Could not parse query parameter %v, all query parameters should be a key and value format", keyAndValue)
		}
		params.Add(keyAndValue[0], keyAndValue[1])
	}

	// Submit request
	resp, err := httpclient.DoRequest(ctx, "PUT", resourceURL, params.Encode(), strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("got error %s", err.Error())
	} else if resp == nil {
		return "", fmt.Errorf("got nil response")
	}

	if resp.Body != nil {
		defer resp.Body.Close()

		// Print the body
		resBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		// Check if error response
		if resp.StatusCode >= 400 && resp.StatusCode <= 600 {
			json.PrintJson(string(resBody))
			return "", fmt.Errorf(resp.Status)
		}

		// 204 is no content, so we will skip it.
		if resp.StatusCode != 204 {
			if !skipAliases {
				aliases.SaveAliasesForResources(string(resBody))
			}
		}

		return string(resBody), nil
	} else {
		return "", nil
	}
}
