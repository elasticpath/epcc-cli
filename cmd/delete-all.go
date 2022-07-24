package cmd

import (
	"context"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/apihelper"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/id"
	"github.com/elasticpath/epcc-cli/external/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
	"reflect"
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
			lastIds := make([][]id.IdableAttributes, 1)
			for {
				resourceURL, err := resources.GenerateUrlViaIdableAttributes(resource.GetCollectionInfo, parentEntityIds)

				if err != nil {
					return err
				}

				params := url.Values{}
				params.Add("page[limit]", "25")

				resp, err := httpclient.DoRequest(context.Background(), "GET", resourceURL, params.Encode(), nil)

				if err != nil {
					return err
				}

				ids, err := apihelper.GetResourceIdsFromHttpResponse(resp)
				resp.Body.Close()

				allIds := make([][]id.IdableAttributes, 0)
				for _, id := range ids {
					allIds = append(allIds, append(parentEntityIds, id))
				}

				min := resource.DeleteEntityInfo.MinResources
				if reflect.DeepEqual(allIds, lastIds) {
					if min == len(lastIds) {
						log.Infof("The minimum number of resources for %s is %d, we have tried to delete %d but couldn't delete them, so we are complete",
							resource.PluralName, min, len(allIds))
					} else if min <= len(lastIds) {
						log.Warnf("The minimum number of resources for %s is %d, we have tried to delete %d currently but seem stuck, so we are done."+
							"Please check to ensure that the resource doesn't require related resources deleted first", resource.PluralName, min, len(allIds))
					} else if min > len(lastIds) {
						log.Warnf("The minimum number of resources for %s is %d, we have tried to delete %d currently but seem stuck, so we are done."+
							"Please check to ensure that the resource doesn't require related resources deleted first", resource.PluralName, min, len(allIds))
					}

					break
				} else {
					lastIds = allIds
				}

				if len(allIds) == 0 {
					log.Infof("Total ids retrieved for %s is %d, we are done", resource.PluralName, len(allIds))
					break
				}

				delPage(resource.DeleteEntityInfo, allIds)
			}
		}

		return aliases.ClearAllAliasesForJsonApiType(resource.JsonApiType)
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
func getParentIds(ctx context.Context, resource resources.Resource) ([][]id.IdableAttributes, error) {

	myEntityIds := make([][]id.IdableAttributes, 0)
	if resource.GetCollectionInfo == nil {
		return myEntityIds, fmt.Errorf("resource %s doesn't support GET collection", resource.PluralName)
	}

	types, err := resources.GetTypesOfVariablesNeeded(resource.GetCollectionInfo.Url)

	if err != nil {
		return myEntityIds, err
	}

	if len(types) == 0 {
		myEntityIds = append(myEntityIds, make([]id.IdableAttributes, 0))
		return myEntityIds, nil
	} else {
		immediateParentType := types[len(types)-1]

		parentResource, ok := resources.GetResourceByName(immediateParentType)

		if !ok {
			return myEntityIds, fmt.Errorf("could not find parent resource %s", immediateParentType)
		}

		return apihelper.GetAllIds(ctx, &parentResource)
	}
}

func delPage(urlInfo *resources.CrudEntityInfo, ids [][]id.IdableAttributes) {
	// Create a wait group to run DELETE in parallel
	wg := sync.WaitGroup{}
	for _, idAttr := range ids {
		wg.Add(1)
		go func(idAttr []id.IdableAttributes) {

			defer wg.Done()
			// Find Resource
			// Replace ids with args in resourceURL
			resourceURL, err := resources.GenerateUrlViaIdableAttributes(urlInfo, idAttr)

			if err != nil {
				return
			}

			// Submit request
			resp, err := httpclient.DoRequest(context.TODO(), "DELETE", resourceURL, "", nil)
			if err != nil {
				return
			}
			defer resp.Body.Close()

		}(idAttr)
	}
	wg.Wait()
}
