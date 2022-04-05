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

		for i := 1; i < len(args); i++ {
			deleteURL := resource.DeleteEntityInfo.Url
			deleteURL = strings.Replace(deleteURL, "%1", args[i], 1)

			// Submit request
			resp, err := httpclient.DoRequest(context.TODO(), "DELETE", deleteURL, "", nil)

			if err != nil {
				log.Println(err)
			}
			defer resp.Body.Close()

			// Print the body
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println(err)
			}
			// Check if error response
			if resp.StatusCode >= 400 && resp.StatusCode <= 600 {
				json.PrintJson(string(body))
				log.Println(resp.Status)
			}
		}
		return nil
	},
}
