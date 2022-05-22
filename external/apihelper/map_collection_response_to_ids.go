package apihelper

import (
	json2 "encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

func GetResourceIdsFromHttpResponse(resp *http.Response) ([]string, error) {

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
					if id, ok := mapValue["id"].(string); ok {
						ids = append(ids, id)
					}
				}
			}
		}
	}
	return ids, nil
}
