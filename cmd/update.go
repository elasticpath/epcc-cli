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
	"strings"
)

var update = &cobra.Command{
	Use:   "update <RESOURCE> [PARENT_ID_1] [PARENT_ID_2] [ID]... <KEY_1> <VAL_1> <KEY_2> <VAL_2>...",
	Short: "Updates an entity of a resource.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Find Resource
		resource, ok := resources.GetResourceByName(args[0])
		if !ok {
			return fmt.Errorf("Could not find resource %s", args[0])
		}

		if resource.CreateEntityInfo == nil {
			return fmt.Errorf("resource %s doesn't support UPDATE", args[0])
		}

		// Count ids in CreateEntity
		resourceURL := resource.UpdateEntityInfo.Url
		idCount, err := resources.GetNumberOfVariablesNeeded(resourceURL)
		if err != nil {
			return err
		}

		// Replace ids with args in resourceURL
		resourceURL, err = resources.GenerateUrl(resourceURL, args[1:])
		if err != nil {
			return err
		}

		args = append(args, "type", resource.JsonApiType)
		// Create the body from remaining args
		body, err := json.ToJson(args[(idCount+1):], noWrapping)
		if err != nil {
			return err
		}

		// Submit request
		resp, err := httpclient.DoRequest(context.TODO(), "PUT", resourceURL, "", strings.NewReader(body))
		if err != nil {
			return fmt.Errorf("Got error %s", err.Error())
		}
		defer resp.Body.Close()

		// Print the body
		resBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		// Check if error response
		if resp.StatusCode >= 400 && resp.StatusCode <= 600 {
			json.PrintJson(string(resBody))
			return fmt.Errorf(resp.Status)
		}

		return json.PrintJson(string(resBody))
	},

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completion.Complete(completion.Request{
				Type: completion.CompleteSingularResource,
			})
		}

		return []string{}, cobra.ShellCompDirectiveNoFileComp
	},
}
