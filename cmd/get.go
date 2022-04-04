package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var get = &cobra.Command{
	Use:   "get [RESOURCE] [ID_1] [ID_2]",
	Short: "Retrieves a single resource.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Set up client to make requests
		client := &http.Client{
			Timeout: time.Second * 10,
		}

		// Find Resource
		resource, ok := resources.Resources[args[0]]
		if !ok {
			return fmt.Errorf("Could not find resource")
		}

		// Count ids in get-collection
		resourceURL := resource.GetCollectionInfo.Url
		idCount := strings.Count(resourceURL, "%")

		// Determine if call should be get-collection or get-entity
		if (idCount-len(args)+1)%2 != 0 {
			idCount += 1
			resourceURL = resource.GetEntityInfo.Url
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
		resourceURL = resourceURL + "?" + params.Encode()
		if (idCount-len(args)+1)%2 != 0 {
			resourceURL = resourceURL + url.QueryEscape(args[len(args)-1])
		}

		// Create the GET request
		req, err := http.NewRequest("GET", Envs.EPCC_API_BASE_URL+resourceURL, nil)
		if err != nil {
			return fmt.Errorf("Got error %s", err.Error())
		}
		req.Header.Set("user-agent", "golang application")
		req.Header.Set("Authorization", "Bearer: 86afed525eab0255c8690223ce02b787707ed38a")

		// Submit request
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("Got error %s", err.Error())
		}
		defer resp.Body.Close()

		// Check if error response
		if resp.StatusCode >= 400 && resp.StatusCode <= 600 {
			return fmt.Errorf(resp.Status)
		}

		// Print the body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		json.PrintJson(string(body))

		return nil
	},
}
