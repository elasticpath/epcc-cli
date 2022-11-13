package apihelper

import (
	"context"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/id"
	"github.com/elasticpath/epcc-cli/external/resources"
	log "github.com/sirupsen/logrus"
	"net/url"
	"reflect"
)

func GetAllIds(ctx context.Context, resource *resources.Resource) ([][]id.IdableAttributes, error) {
	// TODO make this a channel based instead of array based
	// This must be an unbuffered channel since the receiver won't get the channel until after we have sent in some cases.
	//myEntityIds := make(chan<- []string, 1024)
	//defer close(myEntityIds)

	myEntityIds := make([][]id.IdableAttributes, 0)

	if resource == nil {
		myEntityIds = append(myEntityIds, make([]id.IdableAttributes, 0))
		return myEntityIds, nil
	}

	if resource.GetCollectionInfo == nil {
		return myEntityIds, fmt.Errorf("resource %s doesn't support GET collection", resource.PluralName)
	}

	types, err := resources.GetTypesOfVariablesNeeded(resource.GetCollectionInfo.Url)

	if err != nil {
		return myEntityIds, err
	}

	var parentResource *resources.Resource
	if len(types) == 0 {
		parentResource = nil
	} else {
		immediateParentType := types[len(types)-1]

		myParentResource, ok := resources.GetResourceByName(immediateParentType)

		if !ok {
			return myEntityIds, fmt.Errorf("could not find parent resource %s", immediateParentType)
		}

		parentResource = &myParentResource
	}

	myParentEntityIds, err := GetAllIds(ctx, parentResource)
	if err != nil {
		return myEntityIds, err
	}

	// For each parent entity id we need to loop over the entire collection
	for _, parentEntityIds := range myParentEntityIds {

		resourceURL, err := resources.GenerateUrlViaIdableAttributes(resource.GetCollectionInfo, parentEntityIds)

		if err != nil {
			return myEntityIds, err
		}

		lastPageIds := make([]id.IdableAttributes, 125)
		for i := 0; i < 10000; i += 25 {
			params := url.Values{}
			params.Add("page[limit]", "25")
			params.Add("page[offset]", fmt.Sprintf("%d", i))

			resp, err := httpclient.DoRequest(ctx, "GET", resourceURL, params.Encode(), nil)

			if resp != nil && resp.Body != nil {
				defer resp.Body.Close()
			}

			if err != nil {
				return myEntityIds, err
			}

			ids, _, err := GetResourceIdsFromHttpResponse(resp)

			if reflect.DeepEqual(ids, lastPageIds) {
				log.Debugf("Resource %s does not seem to support pagination as we got the exact same set of ids back as the last page... breaking. This might happen if exactly a paginated number of records is returned", resource.PluralName)
				break
			} else {
				lastPageIds = ids
			}

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
