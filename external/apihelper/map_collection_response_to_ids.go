package apihelper

import (
	json2 "encoding/json"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/id"
)

func GetResourceIdsFromHttpResponse(bodyTxt []byte) ([]id.IdableAttributes, int, error) {

	var jsonStruct = map[string]interface{}{}
	err := json2.Unmarshal(bodyTxt, &jsonStruct)
	if err != nil {
		return nil, 0, fmt.Errorf("response for get was not JSON: %w", err)
	}

	// Collect ids from GET Collection output
	var ids []id.IdableAttributes

	totalResources := -1
	if meta, ok := jsonStruct["meta"].(map[string]interface{}); ok {
		if result, ok := meta["results"].(map[string]interface{}); ok {
			if total, ok := result["total"].(float64); ok {
				totalResources = int(total)
			}
		}
	}

	for _, val := range jsonStruct {
		if arrayType, ok := val.([]interface{}); ok {
			for _, value := range arrayType {
				if mapValue, ok := value.(map[string]interface{}); ok {
					match := false

					idAttr := id.IdableAttributes{}
					if id, ok := mapValue["id"].(string); ok {

						match = true
						idAttr.Id = id
					}

					if slug, ok := mapValue["slug"].(string); ok {

						match = true
						idAttr.Slug = slug
					}

					if sku, ok := mapValue["sku"].(string); ok {

						match = true
						idAttr.Sku = sku
					}

					if match {
						ids = append(ids, idAttr)
					}
				}
			}
		}
	}
	return ids, totalResources, nil
}
