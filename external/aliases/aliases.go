package aliases

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strconv"
)

func SaveAliasesForResources(jsonTxt string) {
	var jsonStruct = map[string]interface{}{}
	// TODO what if we get a JSON array or string back
	json.Unmarshal([]byte(jsonTxt), &jsonStruct)

	results := map[string]map[string]string{}
	visitResources(jsonStruct, "", results)

	log.Infof("All aliases: %s", results)
}

func visitResources(data map[string]interface{}, prefix string, results map[string]map[string]string) {

	if typeObj, typeKeyExists := data["type"]; typeKeyExists {
		if idObj, idKeyExists := data["id"]; idKeyExists {
			if typeKeyValue, typeKeyIsString := typeObj.(string); typeKeyIsString {
				if idKeyValue, idKeyIsString := idObj.(string); idKeyIsString {

					aliases := generateAliasesForStruct(typeKeyValue, idKeyValue, data)

					log.Infof("Found a type and id pair %s => %s under prefix %s, aliases %s", typeKeyValue, idKeyValue, prefix, aliases)

					if _, ok := results[typeKeyValue]; !ok {
						results[typeKeyValue] = make(map[string]string)
					}

					for aliasKey, aliasValue := range aliases {
						results[typeKeyValue][aliasKey] = aliasValue
					}
				}
			}
		}
	}

	// Recursively descend over each element
	for key, val := range data {
		if mapType, ok := val.(map[string]interface{}); ok {
			visitResources(mapType, prefix+"."+key, results)
		}

		if arrayType, ok := val.([]interface{}); ok {
			for idx, value := range arrayType {
				if mapValue, ok := value.(map[string]interface{}); ok {
					visitResources(mapValue, prefix+"."+key+"["+strconv.Itoa(idx)+"]", results)
				}

			}
		}

	}

	return
}

func generateAliasesForStruct(typeKey string, idKey string, data map[string]interface{}) map[string]string {
	results := map[string]string{
		// Identity, objects should be an alias of themselves.
		idKey: idKey,
	}

	if alias := getAliasForKey("name", data); alias != "" {
		results[alias] = idKey
	}

	if alias := getAliasForKey("sku", data); alias != "" {
		results[alias] = idKey
	}

	if alias := getAliasForKey("slug", data); alias != "" {
		results[alias] = idKey
	}

	if alias := getAliasForKey("email", data); alias != "" {
		results[alias] = idKey
	}

	if val, ok := data["attributes"]; ok {
		if attributeVal, ok := val.(map[string]interface{}); ok {
			if alias := getAliasForKey("name", attributeVal); alias != "" {
				results[alias] = idKey
			}

			if alias := getAliasForKey("sku", attributeVal); alias != "" {
				results[alias] = idKey
			}

			if alias := getAliasForKey("slug", attributeVal); alias != "" {
				results[alias] = idKey
			}

			if alias := getAliasForKey("email", attributeVal); alias != "" {
				results[alias] = idKey
			}
		}
	}

	return results
}

func getAliasForKey(key string, data map[string]interface{}) string {
	if val, ok := data[key]; ok {
		if strVal, ok := val.(string); ok {

			retVal := fmt.Sprintf("%s-%s", key, strVal)
			return retVal
		} else {
			return ""
		}
	} else {
		return ""
	}
}
