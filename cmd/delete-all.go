package cmd

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"sync"

	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/apihelper"
	"github.com/elasticpath/epcc-cli/external/clictx"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/id"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func NewDeleteAllCommand(parentCmd *cobra.Command) func() {

	var deleteAll = &cobra.Command{
		Use:          "delete-all",
		Short:        "Deletes all of a resource",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("please specify a resource, epcc delete-all [RESOURCE], see epcc delete-all --help")
			} else {
				return fmt.Errorf("invalid resource [%s] specified, see all with epcc delete-all --help", args[0])
			}
		},
	}

	e := config.GetEnv()
	hiddenResources := map[string]struct{}{}

	for _, v := range e.EPCC_CLI_DISABLE_RESOURCES {
		hiddenResources[v] = struct{}{}
	}

	for _, resource := range resources.GetPluralResources() {

		if resource.GetCollectionInfo == nil {
			continue
		}

		if resource.DeleteEntityInfo == nil {
			continue
		}

		if _, ok := hiddenResources[resource.SingularName]; ok {
			log.Tracef("Hiding resource %s", resource.SingularName)
			continue
		}

		if _, ok := hiddenResources[resource.PluralName]; ok {
			log.Tracef("Hiding resource %s", resource.SingularName)
			continue
		}

		resourceName := resource.PluralName

		var deleteAllResourceCmd = &cobra.Command{
			Use:    resourceName,
			Short:  GetDeleteAllShort(resource),
			Hidden: false,
			RunE: func(cmd *cobra.Command, args []string) error {
				return deleteAllInternal(clictx.Ctx, append([]string{resourceName}, args...))
			},
		}
		deleteAll.AddCommand(deleteAllResourceCmd)
	}
	parentCmd.AddCommand(deleteAll)
	return func() {}

}
func deleteAllInternal(ctx context.Context, args []string) error {
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

	allParentEntityIds, err := getParentIds(ctx, resource)

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

			resp, err := httpclient.DoRequest(ctx, "GET", resourceURL, params.Encode(), nil)

			if err != nil {
				return err
			}

			if resp.StatusCode >= 400 {
				log.Warnf("Could not retrieve page of data, aborting")
				break
			}

			ids, totalCount, err := apihelper.GetResourceIdsFromHttpResponse(resp)
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
				log.Infof("Total ids retrieved for %s in %s is %d, we are done", resource.PluralName, resourceURL, len(allIds))
				break
			} else {
				if totalCount >= 0 {
					log.Infof("Total number of %s in %s is %d", resource.PluralName, resourceURL, totalCount)
				} else {
					log.Infof("Total number %s in %s is unknown", resource.PluralName, resourceURL)
				}

			}

			delPage(ctx, resource.DeleteEntityInfo, allIds)
		}
	}

	return aliases.ClearAllAliasesForJsonApiType(resource.JsonApiType)
}

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

var flowsUrlRegex = regexp.MustCompile("^/v2/flows/([^/]+)$")
var cartsUrlRegex = regexp.MustCompile("^/v2/carts/([^/]+)$")

func delPage(ctx context.Context, urlInfo *resources.CrudEntityInfo, ids [][]id.IdableAttributes) {
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
			resp, err := httpclient.DoRequest(ctx, "DELETE", resourceURL, "", nil)
			if err != nil {
				return
			}

			if resp == nil {
				return
			}

			if resp.Body != nil {
				defer resp.Body.Close()
			}
			if resp.StatusCode == 400 && cartsUrlRegex.MatchString(resourceURL) {
				log.Warnf("Could not delete %s, likely associations, trying to clean up", resourceURL)

				resp2, err := httpclient.DoRequest(ctx, "GET", resourceURL, "", nil)

				if err != nil {
					return
				}

				if resp2 != nil {
					if resp2.Body != nil {
						defer resp2.Body.Close()
					}
				}

				if resp2.StatusCode != 200 {
					log.Warnf("Couldn't retrieve cart %d", resp2.StatusCode)
					return
				}

				bytes, err := io.ReadAll(resp2.Body)

				if err != nil {
					log.Warnf("Couldn't read cart body for %s: %v", resourceURL, err)
					return
				}

				custIds, err := json.RunJQOnStringWithArray(".data.relationships.customers.data[].id", string(bytes))

				if err == nil {
					for _, id := range custIds {
						jsonBody := fmt.Sprintf(`{ "data": [{ "id":"%s", "type": "customer"}]}`, id)
						resp3, err := httpclient.DoRequest(ctx, "DELETE", fmt.Sprintf("%s/relationships/customers", resourceURL), "", strings.NewReader(jsonBody))

						if err != nil {
							log.Warnf("Couldn't delete customer cart association %s: %v", id, err)
							continue
						}

						if resp3 == nil {
							continue
						}

						if resp3.Body != nil {
							defer resp3.Body.Close()
						}
					}
				} else {

					// JQ might give us an error if there are no customers (perhaps because there are accounts).
					log.Tracef("Couldn't parse customers, %v", err)
				}

				acctIds, err := json.RunJQOnStringWithArray(".data.relationships.accounts.data[].id", string(bytes))

				if err == nil {
					for _, id := range acctIds {
						jsonBody := fmt.Sprintf(`{ "data": [{ "id":"%s", "type": "account"}]}`, id)
						resp3, err := httpclient.DoRequest(ctx, "DELETE", fmt.Sprintf("%s/relationships/accounts", resourceURL), "", strings.NewReader(jsonBody))

						if err != nil {
							log.Warnf("Couldn't delete account cart association %s: %v", id, err)
							continue
						}

						if resp3 == nil {
							continue
						}

						if resp3.Body != nil {
							defer resp3.Body.Close()
						}
					}
				} else {

					// JQ might give us an error if there are no customers (perhaps because there are accounts).
					log.Tracef("Couldn't parse customers, %v", err)
				}

				resp3, err := httpclient.DoRequest(ctx, "DELETE", resourceURL, "", nil)
				if err != nil {
					return
				}

				if resp3 == nil {
					return
				}

				if resp3.Body != nil {
					defer resp3.Body.Close()
				}

				if resp3.StatusCode == 400 {
					log.Infof("Even after cleaning up associations, still couldn't clean up the cart")
				}
			}

			if resp.StatusCode == 405 && flowsUrlRegex.MatchString(resourceURL) {
				log.Warnf("Could not delete %s, likely core flow, trying to rename and then try again", resourceURL)
				matches := flowsUrlRegex.FindStringSubmatch(resourceURL)

				if len(matches) != 2 {
					log.Errorf("Couldn't get capture group for string [%s], matches %v", resourceURL, matches)
				}

				id := matches[1]
				jsonBody := fmt.Sprintf(`{ "data": { "id":"%s", "type": "flow", "slug": "delete-%s" }}`, id, id)
				resp2, err := httpclient.DoRequest(ctx, "PUT", resourceURL, "", strings.NewReader(jsonBody))

				if err != nil {
					return
				}

				if resp2 != nil {
					if resp.Body != nil {
						defer resp.Body.Close()
					}
				}

				resp3, err := httpclient.DoRequest(ctx, "DELETE", resourceURL, "", nil)
				if err != nil {
					return
				}

				if resp3 == nil {
					return
				}

				if resp3.Body != nil {
					defer resp3.Body.Close()
				}

			}

		}(idAttr)
	}
	wg.Wait()
}
