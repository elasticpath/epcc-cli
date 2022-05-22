package apihelper

import (
	"context"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/resources"
	"net/url"
)

//
func GetAllIds(ctx context.Context, resource *resources.Resource) ([][]string, error) {
	// TODO make this a channel based instead of array based
	// This must be an unbuffered channel since the receiver won't get the channel until after we have sent in some cases.
	//myEntityIds := make(chan<- []string, 1024)
	//defer close(myEntityIds)

	myEntityIds := make([][]string, 0)

	if resource == nil {
		myEntityIds = append(myEntityIds, make([]string, 0))
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

		resourceURL, err := resources.GenerateUrl(*resource, resource.GetCollectionInfo.Url, parentEntityIds)

		if err != nil {
			return myEntityIds, err
		}

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

			ids, err := GetResourceIdsFromHttpResponse(resp)

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
