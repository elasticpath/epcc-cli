package cmd

import (
	json2 "encoding/json"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"sync"
)

var DeleteAll = &cobra.Command{
	Use:    "delete-all [RESOURCE] [ID_1] [ID_2]",
	Short:  "Deletes all of a resource.",
	Args:   cobra.MinimumNArgs(1),
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Find Resource
		resource, ok := resources.GetResourceByName(args[0])
		if !ok {
			return fmt.Errorf("Could not find resource %s", args[0])
		}

		if resource.GetCollectionInfo == nil {
			return fmt.Errorf("Resource %s doesn't support GET collection", args[0])
		}

		if resource.DeleteEntityInfo == nil {
			return fmt.Errorf("Resource %s doesn't support DELETE", args[0])
		}

		min := resource.DeleteEntityInfo.MinResources

		delName := resource.SingularName

		ids, err := getPage(args[0])
		if err != nil {
			return fmt.Errorf("Problem getting page of ids for resource %s", args[0])
		}

		for len(ids) > min {
			delPage(delName, ids)
			ids, err = getPage(args[0])
			if err != nil {
				return fmt.Errorf("Problem getting page of ids for resource %s", args[0])
			}
		}

		return nil
	},

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completion.Complete(completion.Request{
				Type: completion.CompletePluralResource,
				Verb: completion.Delete,
			})
		}

		return []string{}, cobra.ShellCompDirectiveNoFileComp
	},
}

func getPage(resourceName string) ([]string, error) {
	resp, err := getResource([]string{resourceName, "page[limit]", "25"})

	if err != nil {
		return []string{}, err
	}

	// Read the body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var jsonStruct = map[string]interface{}{}
	err = json2.Unmarshal(body, &jsonStruct)
	if err != nil {
		return nil, fmt.Errorf("Response for get was not JSON")
	}

	// Collect ids from GET Collection output
	var ids []string
	for _, val := range jsonStruct {
		if arrayType, ok := val.([]interface{}); ok {
			for _, value := range arrayType {
				if mapValue, ok := value.(map[string]interface{}); ok {
					ids = append(ids, mapValue["id"].(string))
				}
			}
		}
	}
	return ids, nil
}

func delPage(resourceName string, ids []string) {
	// Create a wait group to run DELETE in parallel
	wg := sync.WaitGroup{}
	for _, id := range ids {
		wg.Add(1)
		go func(id string) {
			deleteResource([]string{resourceName, id})
			wg.Done()
		}(id)
	}
	wg.Wait()
}
