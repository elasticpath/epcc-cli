package cmd

import (
	"context"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

var create = &cobra.Command{
	Use:   "create <RESOURCE> [ID_1] [ID_2]... <KEY_1> <VAL_1> <KEY_2> <VAL_2>...",
	Short: "Creates an entity of a resource.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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

		// Submit request
		resp, err := httpclient.DoRequest(context.TODO(), "POST", resourceURL, "", strings.NewReader(body))

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
				Type: completion.CompleteResource,
			})
		}

		return []string{}, cobra.ShellCompDirectiveNoFileComp
	},
}
