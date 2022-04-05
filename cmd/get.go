package cmd

import (
	"context"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
)

var get = &cobra.Command{
	Use:   "get [RESOURCE] [ID_1] [ID_2]",
	Short: "Retrieves a single resource.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Find Resource
		resource, ok := resources.Resources[args[0]]
		if !ok {
			return fmt.Errorf("Could not find resource %s", args[0])
		}

		var resourceURL string
		var idCount int

		if resource.GetCollectionInfo == nil && resource.GetEntityInfo == nil {
			return fmt.Errorf("Resource %s doesn't support GET", args[0])
		} else if resource.GetCollectionInfo != nil && resource.GetEntityInfo == nil {
			resourceURL = resource.GetCollectionInfo.Url
			idCount = strings.Count(resourceURL, "%")
		} else if resource.GetCollectionInfo == nil && resource.GetEntityInfo != nil {
			resourceURL = resource.GetEntityInfo.Url
			idCount = strings.Count(resourceURL, "%")
		} else {
			// Count ids in get-collection
			resourceURL = resource.GetCollectionInfo.Url
			idCount = strings.Count(resourceURL, "%")

			// Determine if call should be get-collection or get-entity
			if (idCount-len(args)+1)%2 != 0 {
				idCount += 1
				resourceURL = resource.GetEntityInfo.Url
			}
		}

		// Replace ids with args in resourceURL
		for i := 1; i <= idCount; i++ {
			resourceURL = strings.Replace(resourceURL, "%"+strconv.Itoa(i), args[i], 1)
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
}
