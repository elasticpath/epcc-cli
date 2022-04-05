package cmd

import (
	"context"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

var delete = &cobra.Command{
	Use:   "delete [RESOURCE] [ID_1] [ID_2]",
	Short: "Deletes a single resource.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Find Resource
		resource, ok := resources.Resources[args[0]]
		if !ok {
			return fmt.Errorf("Could not find resource %s", args[0])
		}

		if resource.DeleteEntityInfo == nil {
			return fmt.Errorf("Resource %s doesn't support DELETE", args[0])
		}

		deleteURL := resource.DeleteEntityInfo.Url
		idCount := strings.Count(deleteURL, "%")
		if len(args)-1 != idCount {
			return fmt.Errorf("Not enough args")
		}

		for i := 0; i < idCount; i++ {
			deleteURL = strings.Replace(deleteURL, "%"+strconv.Itoa(i+1), args[i+1], 1)
		}

		// Submit request
		resp, err := httpclient.DoRequest(context.TODO(), "DELETE", deleteURL, "", nil)
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
			log.Println(resp.Status)
		}

		return json.PrintJson(string(body))
	},
}
