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

var get = &cobra.Command{
	Use:   "get [RESOURCE] [ID_1] [ID_2]",
	Short: "Retrieves a single resource.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body, err := getInternal(args)
		if err != nil {
			return err
		}

		return json.PrintJson(body)
	},

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completion.Complete(completion.Request{
				Type: completion.CompleteSingularResource | completion.CompletePluralResource,
				Verb: completion.Get,
			})
		} else if resource, ok := resources.GetResourceByName(args[0]); ok {
			// len(args) == 0 means complete resource
			// len(args) == 1 means first id
			// lens(args) == 2 means second id.

			resourceURL, err := getUrl(resource, args)
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}

			idCount, err := resources.GetNumberOfVariablesNeeded(resourceURL.Url)

			if err != nil {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}

			if len(args) > 0 && len(args) < 1+idCount {
				// Must be for a resource completion
				types, err := resources.GetTypesOfVariablesNeeded(resourceURL.Url)

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

			} else if len(args) >= idCount+1 { // Arg is after IDs
				if (len(args)-idCount)%2 == 1 { // This is a query param key
					if resource.SingularName != args[0] { // If the resource is plural/get-collection
						return completion.Complete(completion.Request{
							Type:     completion.CompleteQueryParamKey,
							Resource: resource,
							Verb:     completion.GetAll,
						})
					} else {
						return completion.Complete(completion.Request{
							Type:     completion.CompleteQueryParamKey,
							Resource: resource,
							Verb:     completion.Get,
						})
					}
				} else {
					// This is a query param value
					if resource.SingularName != args[0] { // If the resource is plural/get-collection
						return completion.Complete(completion.Request{
							Type:       completion.CompleteQueryParamValue,
							Resource:   resource,
							Verb:       completion.GetAll,
							QueryParam: args[len(args)-1],
						})
					} else {
						return completion.Complete(completion.Request{
							Type:       completion.CompleteQueryParamValue,
							Resource:   resource,
							Verb:       completion.Get,
							QueryParam: args[len(args)-1],
						})
					}
				}
			}
		}

		return []string{}, cobra.ShellCompDirectiveNoFileComp
	},
}

func getInternal(args []string) (string, error) {
	resp, err := getResource(args)

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

		aliases.SaveAliasesForResources(string(body))

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

func getResource(args []string) (*http.Response, error) {
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
	resourceURL, err := resources.GenerateUrl(resourceUrlInfo, args[1:])

	if err != nil {
		return nil, err
	}

	if crud.OverrideUrlPath != "" {
		log.Warnf("Overriding URL Path from %s to %s", resourceURL, crud.OverrideUrlPath)
		resourceURL = crud.OverrideUrlPath
	}

	// Add remaining args as query params
	params := url.Values{}
	for i := idCount + 1; i+1 < len(args); i = i + 2 {
		params.Add(args[i], args[i+1])
	}

	for _, v := range crud.QueryParameters {
		keyAndValue := strings.SplitN(v, "=", 2)
		if len(keyAndValue) != 2 {
			return nil, fmt.Errorf("Could not parse query parameter %v, all query parameters should be a key and value format", keyAndValue)
		}
		params.Add(keyAndValue[0], keyAndValue[1])
	}

	// Steve doesn't understand this logic check
	if (idCount-len(args)+1)%2 != 0 {
		resourceURL = resourceURL + url.QueryEscape(args[len(args)-1])
	}

	// Submit request
	resp, err := httpclient.DoRequest(context.TODO(), "GET", resourceURL, params.Encode(), nil)

	if err != nil {
		return nil, fmt.Errorf("got error %s", err.Error())
	}

	return resp, nil
}
