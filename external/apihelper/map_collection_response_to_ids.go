package apihelper

import (
	json2 "encoding/json"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/id"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

func GetResourceIdsFromHttpResponse(resp *http.Response) ([]id.IdableAttributes, error) {

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
	var ids []id.IdableAttributes
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
	return ids, nil
}
