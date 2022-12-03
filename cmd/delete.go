package cmd

import (
	"context"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/crud"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var delete = &cobra.Command{
	Use:   "delete [RESOURCE] [ID_1] [ID_2]",
	Short: "Deletes a single resource.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		body, err := deleteInternal(context.Background(), args)
		if err != nil {
			return err
		}

		return json.PrintJson(body)
	},

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completion.Complete(completion.Request{
				Type: completion.CompleteSingularResource,
				Verb: completion.Delete,
			})
		} else if resource, ok := resources.GetResourceByName(args[0]); ok {
			// len(args) == 0 means complete resource
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

			if len(args) > 0 && len(args) < 1+idCount {
				// Must be for a resource completion
				types, err := resources.GetTypesOfVariablesNeeded(resource.DeleteEntityInfo.Url)

				if err != nil {
					return []string{}, cobra.ShellCompDirectiveNoFileComp
				}

				typeIdxNeeded := len(args) - 1

				if completionResource, ok := resources.GetResourceByName(types[typeIdxNeeded]); ok {
					return completion.Complete(completion.Request{
						Type:     completion.CompleteAlias,
						Resource: completionResource,
					})
				}
			} else {
				if (len(args)-idCount)%2 == 1 { // This is an attribute key
					usedAttributes := make(map[string]int)
					for i := idCount + 1; i < len(args); i = i + 2 {
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
						Type:      completion.CompleteAttributeValue,
						Resource:  resource,
						Verb:      completion.Delete,
						Attribute: args[len(args)-1],
					})
				}
			}
		}

		return []string{}, cobra.ShellCompDirectiveNoFileComp
	},
}

func deleteInternal(ctx context.Context, args []string) (string, error) {
	crud.OutstandingRequestCounter.Add(1)
	defer crud.OutstandingRequestCounter.Done()

	resource, ok := resources.GetResourceByName(args[0])
	if !ok {
		return "", fmt.Errorf("could not find resource %s", args[0])
	}

	resp, err := deleteResource(ctx, args)
	if err != nil {
		return "", err
	}

	if resp == nil {
		return "", fmt.Errorf("got nil response")
	}

	idToDelete := aliases.ResolveAliasValuesOrReturnIdentity(resource.JsonApiType, resource.AlternateJsonApiTypesForAliases, "id", args[len(args)-1])
	aliases.DeleteAliasesById(idToDelete, resource.JsonApiType)

	if resp.Body != nil {

		defer resp.Body.Close()

		// Print the body
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			log.Fatal(err)
		}

		return string(body), nil
	} else {
		return "", nil
	}

}

func deleteResource(ctx context.Context, args []string) (*http.Response, error) {
	// Find Resource
	resource, ok := resources.GetResourceByName(args[0])
	if !ok {
		return nil, fmt.Errorf("could not find resource %s", args[0])
	}

	if resource.DeleteEntityInfo == nil {
		return nil, fmt.Errorf("resource %s doesn't support DELETE", args[0])
	}

	// Replace ids with args in resourceURL
	resourceURL, err := resources.GenerateUrl(resource.DeleteEntityInfo, args[1:])

	if err != nil {
		return nil, err
	}

	if crud.OverrideUrlPath != "" {
		log.Warnf("Overriding URL Path from %s to %s", resourceURL, crud.OverrideUrlPath)
		resourceURL = crud.OverrideUrlPath
	}

	params := url.Values{}

	for _, v := range crud.QueryParameters {
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
		body, err := json.ToJson(jsonArgs, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes)

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
