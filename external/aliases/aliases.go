package aliases

import (
	"encoding/json"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/profiles"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// We will serialize our access for aliases to prevent parallel operations in the same process from losing data
// however, we should use file locking in the OS to stop multiple concurrent invocations.
var filelock = sync.Mutex{}

func GetAliasDataDirectory() string {
	profileDirectory := profiles.GetProfileDataDirectory()
	profileDataDirectory := filepath.FromSlash(profileDirectory + "/aliases/")

	//built in check if dir exists
	if err := os.MkdirAll(profileDataDirectory, 0700); err != nil {
		log.Errorf("could not make directory")
	}

	return profileDataDirectory
}
func GetAliasesForJsonApiType(jsonApiType string) map[string]string {
	profileDirectory := GetAliasDataDirectory()
	aliasFile := getAliasFileForJsonApiType(profileDirectory, jsonApiType)

	aliasMap := map[string]string{}

	data, err := ioutil.ReadFile(aliasFile)
	if err != nil {
		log.Debugf("Could not read %s, error %s", aliasFile, err)
		data = []byte{}
	} else {
	}

	err = yaml.Unmarshal(data, aliasMap)
	if err != nil {
		log.Debugf("Could not unmarshall existing file %s, error %s", data, err)
	}

	return aliasMap
}

func ResolveAliasValuesOrReturnIdentity(jsonApiType string, value string) string {
	if result, ok := GetAliasesForJsonApiType(jsonApiType)[value]; ok {
		return result
	}
	return value
}

func SaveAliasesForResources(jsonTxt string) {
	var jsonStruct = map[string]interface{}{}
	err := json.Unmarshal([]byte(jsonTxt), &jsonStruct)
	if err != nil {
		log.Warnf("Response was not JSON so not scanning for aliases")
		return
	}

	results := map[string]map[string]string{}
	visitResources(jsonStruct, "", results)

	log.Tracef("All aliases: %s", results)

	profileDirectory := GetAliasDataDirectory()
	for resourceType, aliases := range results {
		saveAliasesForResource(profileDirectory, resourceType, aliases)
	}

}

// This function saves all the aliases for a specific resource.
func saveAliasesForResource(profileDirectory string, jsonApiType string, aliases map[string]string) {
	filelock.Lock()
	defer filelock.Unlock()

	aliasFile := getAliasFileForJsonApiType(profileDirectory, jsonApiType)
	data, err := ioutil.ReadFile(aliasFile)
	if err != nil {
		log.Debugf("Could not read %s, error %s", aliasFile, err)
		data = []byte{}
	} else {
	}

	aliasMap := map[string]string{}

	err = yaml.Unmarshal(data, aliasMap)
	if err != nil {
		log.Debugf("Could not unmarshall existing file %s, error %s", data, err)
	}

	for key, value := range aliases {
		key0 := strings.Split(key, "=")[0]
		for oldKey, oldValue := range aliasMap {
			oldKey0 := strings.Split(oldKey, "=")[0]
			oldKey1 := strings.Split(oldKey, "=")[1]
			if oldValue == value && oldKey0 == key0 {
				delete(aliasMap, oldKey0+"="+oldKey1)
			}
		}
	}

	for key, value := range aliases {
		aliasMap[key] = value
	}

	// We will write to a temp file and then rename, to prevent data loss. rename's in the same folder are likely atomic in most settings.
	// Although we should probably sync on the file as well, that might be too much overhead, and I was too lazy to rewrite this
	// https://github.com/golang/go/issues/20599
	tmpFileName := aliasFile + "." + uuid.New().String()

	marshal, err := yaml.Marshal(aliasMap)

	if err != nil {
		log.Warnf("Could not save aliases for %s, error %v", tmpFileName, err)
	}

	err = ioutil.WriteFile(tmpFileName, marshal, 0600)
	if err != nil {
		log.Warnf("Could not save aliases for %s, error %v", tmpFileName, err)
	}

	err = os.Rename(tmpFileName, aliasFile)
	if err != nil {
		log.Warnf("Could not save aliases for %s, error %v", tmpFileName, err)
	}
}

func getAliasFileForJsonApiType(profileDirectory string, resourceType string) string {
	aliasFile := fmt.Sprintf("%s/aliases_%s.yml", profileDirectory, resourceType)
	return aliasFile
}

func visitResources(data map[string]interface{}, prefix string, results map[string]map[string]string) {

	if typeObj, typeKeyExists := data["type"]; typeKeyExists {
		if idObj, idKeyExists := data["id"]; idKeyExists {
			if typeKeyValue, typeKeyIsString := typeObj.(string); typeKeyIsString {
				if idKeyValue, idKeyIsString := idObj.(string); idKeyIsString {
					aliases := generateAliasesForStruct(typeKeyValue, idKeyValue, data)

					log.Tracef("Found a type and id pair %s => %s under prefix %s, aliases %s", typeKeyValue, idKeyValue, prefix, aliases)

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
		"id=" + idKey: idKey,
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
			// TODO not sure why spaces don't work with auto completes.
			retVal := strings.ReplaceAll(fmt.Sprintf("%s=%s", key, strVal), " ", "_")
			return retVal
		} else {
			return ""
		}
	} else {
		return ""
	}
}
