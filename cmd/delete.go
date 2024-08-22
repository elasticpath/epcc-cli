package cmd

import (
	"context"
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
	}

	for _, resource := range resources.GetPluralResources() {
		if resource.DeleteEntityInfo == nil {
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

					body, err := deleteInternal(context.Background(), overrides, allow404, append([]string{resourceName}, args...))

					if err != nil {
						if body != "" {
							if !noBodyPrint {
								json.PrintJsonToStdout(body)
							}
						}
						return err
					}

					if noBodyPrint {
						return nil
					} else {
						return json.PrintJsonToStdout(body)
					}
				}

				return repeater(c, repeat, repeatDelay, cmd, args, ignoreErrors)
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

	parentCmd.AddCommand(deleteCmd)

	return resetFunc
}
func deleteInternal(ctx context.Context, overrides *httpclient.HttpParameterOverrides, allow404 bool, args []string) (string, error) {
	shutdown.OutstandingOpCounter.Add(1)
	defer shutdown.OutstandingOpCounter.Done()

	resource, ok := resources.GetResourceByName(args[0])
	if !ok {
		return "", fmt.Errorf("could not find resource %s", args[0])
	}

	resp, err := deleteResource(ctx, overrides, args)
	if err != nil {
		return "", err
	}

	if resp == nil {
		return "", fmt.Errorf("got nil response")
	}

	if resp.StatusCode < 400 {
		idToDelete := aliases.ResolveAliasValuesOrReturnIdentity(resource.JsonApiType, resource.AlternateJsonApiTypesForAliases, args[len(args)-1], "id")
		aliases.DeleteAliasesById(idToDelete, resource.JsonApiType)
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
			if resp.StatusCode != 404 || !allow404 {
				return string(body), fmt.Errorf(resp.Status)
			}
		}

		return string(body), nil
	} else {
		return "", nil
	}

}

func deleteResource(ctx context.Context, overrides *httpclient.HttpParameterOverrides, args []string) (*http.Response, error) {
	// Find Resource
	resource, ok := resources.GetResourceByName(args[0])
	if !ok {
		return nil, fmt.Errorf("could not find resource %s", args[0])
	}

	if resource.DeleteEntityInfo == nil {
		return nil, fmt.Errorf("resource %s doesn't support DELETE", args[0])
	}

	// Replace ids with args in resourceURL
	resourceURL, err := resources.GenerateUrl(resource.DeleteEntityInfo, args[1:], true)

	if err != nil {
		return nil, err
	}

	if overrides.OverrideUrlPath != "" {
		log.Warnf("Overriding URL Path from %s to %s", resourceURL, overrides.OverrideUrlPath)
		resourceURL = overrides.OverrideUrlPath
	}

	params := url.Values{}

	for _, v := range overrides.QueryParameters {
		keyAndValue := strings.SplitN(v, "=", 2)
		if len(keyAndValue) != 2 {
			return nil, fmt.Errorf("Could not parse query parameter %v, all query parameters should be a key and value format", keyAndValue)
		}
		params.Add(keyAndValue[0], keyAndValue[1])
	}

	idCount, err := resources.GetNumberOfVariablesNeeded(resource.DeleteEntityInfo.Url)

	if !resource.NoWrapping {
		args = append(args, "type", resource.JsonApiType)
	}
	// Create the body from remaining args

	jsonArgs := args[(idCount + 1):]

	var payload io.Reader = nil
	if len(jsonArgs) > 0 {
		body, err := json.ToJson(jsonArgs, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes, true)

		if err != nil {
			return nil, err
		}

		payload = strings.NewReader(body)
	}

	// Submit request
	resp, err := httpclient.DoRequest(ctx, "DELETE", resourceURL, params.Encode(), payload)
	if err != nil {
		return nil, fmt.Errorf("got error %s", err.Error())
	}

	return resp, nil
}
