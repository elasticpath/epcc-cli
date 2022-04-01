package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net/http"
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
		url := resource.GetCollectionInfo.Url
		idCount := strings.Count(url, "%")

		// Determine if call should be get-collection or get-entity
		if (idCount-len(args)+1)%2 != 0 {
			idCount += 1
			url = resource.GetEntityInfo.Url
		}

		// Replace ids with args in url
		for i := 1; i <= idCount; i++ {
			url = strings.Replace(url, "%"+strconv.Itoa(i), args[i], 1)
		}

		// Create the GET request
		req, err := http.NewRequest("GET", Envs.EPCC_API_BASE_URL+url, nil)
		if err != nil {
			return fmt.Errorf("Got error %s", err.Error())
		}
		req.Header.Set("user-agent", "golang application")
		req.Header.Set("Authorization", "Bearer: 951a9a5db7b712ce999c3ccdac889805071c1ae7")

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

		fmt.Println(string(body))

		return nil
	},
}
