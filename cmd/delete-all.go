package cmd

import (
	"context"
	json2 "encoding/json"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
)

var DeleteAll = &cobra.Command{
	Use:    "delete-all [RESOURCE]",
	Short:  "Deletes all of a resource.",
	Args:   cobra.MinimumNArgs(1),
	Hidden: false,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Find Resource
		resource, ok := resources.GetResourceByName(args[0])
		if !ok {
			return fmt.Errorf("could not find resource %s", args[0])
		}

		if resource.GetCollectionInfo == nil {
			return fmt.Errorf("resource %s doesn't support GET collection", args[0])
		}

		if resource.DeleteEntityInfo == nil {
			return fmt.Errorf("resource %s doesn't support DELETE", args[0])
		}

		allParentEntityIds, err := getParentIds(context.Background(), resource)

		if err != nil {
			return fmt.Errorf("could not retrieve parent ids for for resource %s, error: %w", resource.PluralName, err)
		}

		if len(allParentEntityIds) == 1 {
			log.Infof("Resource %s is a top level resource need to scan only one path to delete all resources", resource.PluralName)
		} else {
			log.Infof("Resource %s is not a top level resource, need to scan %d paths to delete all resources", resource.PluralName, len(allParentEntityIds))
		}

		for _, parentEntityIds := range allParentEntityIds {
			for ; ; {
				resourceURL, err := resources.GenerateUrl(resource, resource.GetCollectionInfo.Url, parentEntityIds)

				if err != nil {
					return err
				}

				params := url.Values{}
				params.Add("page[limit]", "25")

				resp, err := httpclient.DoRequest(context.Background(), "GET", resourceURL, params.Encode(), nil)

				if err != nil {
					return err
				}

				ids, err := getResourceIdsFromHttpResponse(resp)
				resp.Body.Close()

				min := resource.DeleteEntityInfo.MinResources

				if len(ids) <= min {
					break
				}

				allIds := make([][]string, 0)
				for _, id := range ids {
					allIds = append(allIds, append(parentEntityIds, id))

				}

				delPage(resource.PluralName, allIds)
			}
		}

		/*
			min := resource.DeleteEntityInfo.MinResources

			delName := resource.SingularName

			ids, err := getPageOfResourcesToDelete(args[0])
			if err != nil {
				return fmt.Errorf("problem getting page of ids for resource %s", args[0])
			}

			for len(ids) > min {
				delPage(delName, ids)
				ids, err = getPageOfResourcesToDelete(args[0])
				if err != nil {
					return fmt.Errorf("problem getting page of ids for resource %s", args[0])
				}
			}

		*/

		//os.Remove(aliases.GetAliasFileForJsonApiType(aliases.GetAliasDataDirectory(), resource.JsonApiType))

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

//
func getParentIds(ctx context.Context, resource resources.Resource) ([][]string, error) {
	// TODO make this a channel based instead of array based
	// This must be an unbuffered channel since the receiver won't get the channel until after we have sent in some cases.
	//myEntityIds := make(chan<- []string, 1024)
	//defer close(myEntityIds)

	myEntityIds := make([][]string, 0)
	if resource.GetCollectionInfo == nil {
		return myEntityIds, fmt.Errorf("resource %s doesn't support GET collection", resource.PluralName)
	}

	types, err := resources.GetTypesOfVariablesNeeded(resource.GetCollectionInfo.Url)

	if err != nil {
		return myEntityIds, err
	}

	if len(types) == 0 {
		myEntityIds = append(myEntityIds, make([]string, 0))
		return myEntityIds, nil
	} else {
		immediateParentType := types[len(types)-1]

		parentResource, ok := resources.GetResourceByName(immediateParentType)

		if !ok {
			return myEntityIds, fmt.Errorf("could not find parent resource %s", immediateParentType)
		}

		myParentEntityIds, err := getParentIds(ctx, parentResource)

		if err != nil {
			return myEntityIds, err
		}

		// For each parent entity id we need to loop over the entire collection
		for _, parentEntityIds := range myParentEntityIds {

			resourceURL, err := resources.GenerateUrl(resource, parentResource.GetCollectionInfo.Url, parentEntityIds)

			if err != nil {
				return myEntityIds, err
			}

			for i := 0; i < 10000; i += 25 {
				params := url.Values{}
				params.Add("page[limit]", "25")
				params.Add("page[offset]", fmt.Sprintf("%d", i))

				resp, err := httpclient.DoRequest(ctx, "GET", resourceURL, params.Encode(), nil)
				defer resp.Body.Close()

				if err != nil {
					return myEntityIds, err
				}

				ids, err := getResourceIdsFromHttpResponse(resp)

				if len(ids) == 0 {
					break
				}

				if err != nil {
					return myEntityIds, err
				}

				for _, parentId := range ids {
					myEntityIds = append(myEntityIds, append(parentEntityIds, parentId))
				}
			}
		}

		return myEntityIds, nil

	}
}

func getResourceIdsFromHttpResponse(resp *http.Response) ([]string, error) {

	// Read the body
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	var jsonStruct = map[string]interface{}{}
	err = json2.Unmarshal(body, &jsonStruct)
	if err != nil {
		return nil, fmt.Errorf("response for get was not JSON: %w", err)
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

func delPage(resourceName string, ids [][]string) {
	// Create a wait group to run DELETE in parallel
	wg := sync.WaitGroup{}
	for _, id := range ids {
		wg.Add(1)
		go func(id []string) {
			args := make([]string, 0)
			args = append(args, resourceName)
			args = append(args, id...)

			deleteResource(args)
			wg.Done()
		}(id)
	}
	wg.Wait()
}
