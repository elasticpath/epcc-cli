package cmd

import (
	"context"
	gojson "encoding/json"
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
	"net/url"
	"strings"
)

var update = &cobra.Command{
	Use:   "update <RESOURCE> [PARENT_ID_1] [PARENT_ID_2] [ID]... <KEY_1> <VAL_1> <KEY_2> <VAL_2>...",
	Short: "Updates an entity of a resource.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		body, err := updateInternal(context.Background(), args)

		if err != nil {
			return err
		}

		if outputJq != "" {
			output, err := json.RunJQOnStringWithArray(outputJq, body)

			if err != nil {
				return err
			}

			outputJson, err := gojson.Marshal(output)

			if err != nil {
				return err
			}

			return json.PrintJson(string(outputJson))
		}

		return json.PrintJson(body)
	},

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completion.Complete(completion.Request{
				Type: completion.CompleteSingularResource,
				Verb: completion.Update,
			})
		}

		// Find Resource
		resource, ok := resources.GetResourceByName(args[0])
		if ok {
			if resource.UpdateEntityInfo != nil {
				resourceURL := resource.UpdateEntityInfo.Url
				idCount, _ := resources.GetNumberOfVariablesNeeded(resourceURL)
				if len(args)-idCount >= 1 { // Arg is after IDs
					if (len(args)-idCount)%2 == 1 { // This is an attribute key
						usedAttributes := make(map[string]int)
						for i := idCount + 1; i < len(args); i = i + 2 {
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
							Type:       completion.CompleteAttributeValue,
							Resource:   resource,
							Verb:       completion.Update,
							Attribute:  args[len(args)-1],
							ToComplete: toComplete,
						})
					}
				} else {
					// Arg is in IDS
					// Must be for a resource completion
					types, err := resources.GetTypesOfVariablesNeeded(resourceURL)

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
				}
			}
		}

		return []string{}, cobra.ShellCompDirectiveNoFileComp
	},
}

func updateInternal(ctx context.Context, args []string) (string, error) {
	crud.OutstandingRequestCounter.Add(1)
	defer crud.OutstandingRequestCounter.Done()

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
	resourceURL, err := resources.GenerateUrl(resourceUrlInfo, args[1:])
	if err != nil {
		return "", err
	}

	if crud.OverrideUrlPath != "" {
		log.Warnf("Overriding URL Path from %s to %s", resourceURL, crud.OverrideUrlPath)
		resourceURL = crud.OverrideUrlPath
	}

	args = append(args, "type", resource.JsonApiType)
	// Create the body from remaining args
	body, err := json.ToJson(args[(idCount+1):], resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes)
	if err != nil {
		return "", err
	}

	params := url.Values{}

	for _, v := range crud.QueryParameters {
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
			aliases.SaveAliasesForResources(string(resBody))
		}

		return string(resBody), nil
	} else {
		return "", nil
	}
}
