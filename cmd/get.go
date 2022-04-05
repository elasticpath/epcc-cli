package cmd

import (
	"context"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/url"
)

var get = &cobra.Command{
	Use:   "get [RESOURCE] [ID_1] [ID_2]",
	Short: "Retrieves a single resource.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Find Resource
		resource, ok := resources.GetResourceByName(args[0])
		if !ok {
			return fmt.Errorf("Could not find resource %s", args[0])
		}

		var resourceURL string
		var idCount int

		if resource.GetCollectionInfo == nil && resource.GetEntityInfo == nil {
			return fmt.Errorf("Resource %s doesn't support GET", args[0])
		} else if resource.GetCollectionInfo != nil && resource.GetEntityInfo == nil {
			resourceURL = resource.GetCollectionInfo.Url

		} else if resource.GetCollectionInfo == nil && resource.GetEntityInfo != nil {
			resourceURL = resource.GetEntityInfo.Url

		} else {
			// Count ids in get-collection
			resourceURL = resource.GetCollectionInfo.Url
		}

		idCount, err := resources.GetNumberOfVariablesNeeded(resourceURL)

		if err != nil {
			return err
		}

		// Replace ids with args in resourceURL
		resourceURL, err = resources.GenerateUrl(resourceURL, args[1:])

		if err != nil {
			return err
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
			return fmt.Errorf("Got error %s", err.Error())
		}
		defer resp.Body.Close()

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

		return json.PrintJson(string(body))
	},

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completion.Complete(completion.Request{
				Type: completion.CompleteResource,
			})
		}

		return []string{}, cobra.ShellCompDirectiveNoFileComp
	},
}
