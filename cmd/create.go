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

var create = &cobra.Command{
	Use:   "create <RESOURCE> [ID_1] [ID_2]... <KEY_1> <VAL_1> <KEY_2> <VAL_2>...",
	Short: "Creates an entity of a resource.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Find Resource
		resource, ok := resources.GetResourceByName(args[0])
		if !ok {
			return fmt.Errorf("Could not find resource %s", args[0])
		}

		if resource.CreateEntityInfo == nil {
			return fmt.Errorf("resource %s doesn't support CREATE", args[0])
		}

		// Count ids in CreateEntity
		resourceURL := resource.CreateEntityInfo.Url

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
				Type: completion.CompleteSingularResource,
				Verb: completion.Create,
			})
		}

		// Find Resource
		resource, ok := resources.GetResourceByName(args[0])
		if ok {
			if resource.CreateEntityInfo != nil {
				idCount, _ := resources.GetNumberOfVariablesNeeded(resource.CreateEntityInfo.Url)
				if len(args)-idCount >= 1 { // Arg is after IDs
					if (len(args)-idCount)%2 == 1 { // This is an attribute key
						usedAttributes := make(map[string]int)
						for i := idCount + 1; i < len(args); i = i + 2 {
							usedAttributes[args[i]] = 0
						}
						return completion.Complete(completion.Request{
							Type:       completion.CompleteAttributeKey,
							Resource:   resource,
							Attributes: usedAttributes,
							Verb:       completion.Create,
						})
					}
				}
			}
		}

		return []string{}, cobra.ShellCompDirectiveNoFileComp
	},
}
