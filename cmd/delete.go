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
)

var delete = &cobra.Command{
	Use:   "delete [RESOURCE] [ID_1] [ID_2]",
	Short: "Deletes a single resource.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Find Resource
		resource, ok := resources.GetResourceByName(args[0])
		if !ok {
			return fmt.Errorf("Could not find resource %s", args[0])
		}

		if resource.DeleteEntityInfo == nil {
			return fmt.Errorf("Resource %s doesn't support DELETE", args[0])
		}

		resourceURL := resource.DeleteEntityInfo.Url

		// Replace ids with args in resourceURL
		resourceURL, err := resources.GenerateUrl(resourceURL, args[1:])

		if err != nil {
			return err
		}

		// Submit request
		resp, err := httpclient.DoRequest(context.TODO(), "DELETE", resourceURL, "", nil)
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

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completion.Complete(completion.Request{
				Type: completion.CompleteResource,
			})
		}

		return []string{}, cobra.ShellCompDirectiveNoFileComp
	},
}
