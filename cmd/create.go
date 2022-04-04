package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var create = &cobra.Command{
	Use:   "create <RESOURCE> [ID_1] [ID_2]... <KEY_1> <VAL_1> <KEY_2> <VAL_2>...",
	Short: "Creates an entity of a resource.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Set up client to make requests
		client := &http.Client{
			Timeout: time.Second * 10,
		}

		// Find Resource
		resource, ok := resources.Resources[args[0]]
		if !ok {
			return fmt.Errorf("Could not find resource %s", args[0])
		}

		if resource.CreateEntityInfo == nil {
			return fmt.Errorf("resource %s doesn't support CREATE", args[0])
		}

		// Count ids in CreateEntity
		resourceURL := resource.CreateEntityInfo.Url
		idCount := strings.Count(resourceURL, "%")

		// Replace ids with args in resourceURL
		for i := 1; i <= idCount; i++ {
			resourceURL = strings.Replace(resourceURL, "%"+strconv.Itoa(i), args[i], 1)
		}

		args = append(args, "type", resource.JsonApiType)
		// Create the body from remaining args
		body, err := json.ToJson(args[(idCount+1):], noWrapping)

		if err != nil {
			return err
		}

		// Create the CREATE request
		req, err := http.NewRequest("POST", Envs.EPCC_API_BASE_URL+resourceURL, strings.NewReader(body))
		if err != nil {
			return fmt.Errorf("Got error %s", err.Error())
		}
		req.Header.Set("user-agent", "golang application")
		req.Header.Set("content-type", resource.CreateEntityInfo.ContentType)
		req.Header.Set("Authorization", "Bearer: a9503e216234a78913f0545aa6b4209f5569e378")

		// Submit request
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("Got error %s", err.Error())
		}
		defer resp.Body.Close()

		// Print the body
		resBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		json.PrintJson(string(resBody))

		// Check if error response
		if resp.StatusCode >= 400 && resp.StatusCode <= 600 {
			return fmt.Errorf(resp.Status)
		}

		return nil
	},
}
