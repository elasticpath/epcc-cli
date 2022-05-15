package cmd

import (
	"context"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"net/url"
)

var get = &cobra.Command{
	Use:   "get [RESOURCE] [ID_1] [ID_2]",
	Short: "Retrieves a single resource.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := getResource(args)

		if err != nil {
			return err
		}

		// Print the body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		// Check if error response
		if resp.StatusCode >= 400 && resp.StatusCode <= 600 {
			json.PrintJson(string(body))
			return fmt.Errorf(resp.Status)
		}

		aliases.SaveAliasesForResources(string(body))
		return json.PrintJson(string(body))
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

			idCount, err := resources.GetNumberOfVariablesNeeded(resourceURL)

			if err != nil {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}

			if len(args) > 0 && len(args) < 1+idCount {
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

			} else if len(args) >= idCount+1 { // Arg is after IDs
				if (len(args)-idCount)%2 == 1 { // This is a query param key
					if resource.SingularName != args[0] { // If the resource is plural/get-collection
						return completion.Complete(completion.Request{
							Type:     completion.CompleteQueryParam,
							Resource: resource,
							Verb:     completion.GetAll,
						})
					} else {
						return completion.Complete(completion.Request{
							Type:     completion.CompleteQueryParam,
							Resource: resource,
							Verb:     completion.Get,
						})
					}
				}
			}
		}

		return []string{}, cobra.ShellCompDirectiveNoFileComp
	},
}

func getUrl(resource resources.Resource, args []string) (string, error) {
	resourceURL := ""
	if resource.GetCollectionInfo == nil && resource.GetEntityInfo == nil {
		return "", fmt.Errorf("resource %s doesn't support GET", args[0])
	} else if resource.GetCollectionInfo != nil && resource.GetEntityInfo == nil {
		resourceURL = resource.GetCollectionInfo.Url
	} else if resource.GetCollectionInfo == nil && resource.GetEntityInfo != nil {
		resourceURL = resource.GetEntityInfo.Url
	} else {
		if _, ok := resources.GetPluralResources()[args[0]]; ok {
			resourceURL = resource.GetCollectionInfo.Url
		} else {
			resourceURL = resource.GetEntityInfo.Url
		}
	}
	return resourceURL, nil
}

func getResource(args []string) (*http.Response, error) {
	// Find Resource
	resource, ok := resources.GetResourceByName(args[0])
	if !ok {
		return nil, fmt.Errorf("could not find resource %s", args[0])
	}

	var resourceURL string
	var idCount int

	resourceURL, err2 := getUrl(resource, args)
	if err2 != nil {
		return nil, err2
	}

	idCount, err := resources.GetNumberOfVariablesNeeded(resourceURL)

	if err != nil {
		return nil, err
	}

	// Replace ids with args in resourceURL
	resourceURL, err = resources.GenerateUrl(resource, resourceURL, args[1:])

	if err != nil {
		return nil, err
	}

	// Add remaining args as query params
	params := url.Values{}
	for i := idCount + 1; i+1 < len(args); i = i + 2 {
		params.Add(args[i], args[i+1])
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
	defer resp.Body.Close()

	return resp, nil
}
