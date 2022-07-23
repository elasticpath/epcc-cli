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

var aliasDirectoryOverride = ""

type AliasAttributes struct {
	Id   string `yaml:"id"`
	Slug string `yaml:"slug"`
	Sku  string `yaml:"sku"`
}

func ClearAllAliases() error {
	aliasDataDirectory := getAliasDataDirectory()

	if err := os.RemoveAll(aliasDataDirectory); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil

}

func ClearAllAliasesForJsonApiType(jsonApiType string) error {
	if err := os.Remove(getAliasFileForJsonApiType(getAliasDataDirectory(), jsonApiType)); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil

}

func GetAliasesForJsonApiType(jsonApiType string) map[string]*AliasAttributes {
	profileDirectory := getAliasDataDirectory()
	aliasFile := getAliasFileForJsonApiType(profileDirectory, jsonApiType)

	aliasMap := map[string]*AliasAttributes{}

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
		return result.Id
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

	results := map[string]map[string]*AliasAttributes{}
	visitResources(jsonStruct, "", results)

	log.Tracef("All aliases: %v", results)

	for resourceType, aliases := range results {
		saveAliasesForResource(resourceType, aliases)
	}

}

func DeleteAliasesById(id string, jsonApiType string) {
	modifyAliases(jsonApiType, func(m map[string]*AliasAttributes) {
		for key, value := range m {
			if value.Id == id {
				delete(m, key)
			}
		}
	},
	)

}

func getAliasDataDirectory() string {
	aliasDirectory := aliasDirectoryOverride

	if aliasDirectory == "" {
		profileDirectory := profiles.GetProfileDataDirectory()
		profileDataDirectory := filepath.FromSlash(profileDirectory + "/aliases/")
		aliasDirectory = profileDataDirectory
	}

	//built in check if dir exists
	if err := os.MkdirAll(aliasDirectory, 0700); err != nil {
		log.Errorf("could not make directory")
	}

	return aliasDirectory
}

func getAliasFileForJsonApiType(profileDirectory string, resourceType string) string {
	aliasFile := fmt.Sprintf("%s/aliases_%s.yml", profileDirectory, resourceType)
	return aliasFile
}

func modifyAliases(jsonApiType string, fn func(map[string]*AliasAttributes)) map[string]*AliasAttributes {
	profileDirectory := getAliasDataDirectory()
	filelock.Lock()
	defer filelock.Unlock()

	aliasFile := getAliasFileForJsonApiType(profileDirectory, jsonApiType)
	data, err := ioutil.ReadFile(aliasFile)
	if err != nil {
		log.Debugf("Could not read %s, error %s", aliasFile, err)
		data = []byte{}
	}

	existingAliases := map[string]*AliasAttributes{}

	err = yaml.Unmarshal(data, existingAliases)
	if err != nil {
		log.Debugf("Could not unmarshall existing file %s, error %s", data, err)
	}
	fn(existingAliases)
	// We will write to a temp file and then rename, to prevent data loss. rename's in the same folder are likely atomic in most settings.
	// Although we should probably sync on the file as well, that might be too much overhead, and I was too lazy to rewrite this
	// https://github.com/golang/go/issues/20599
	tmpFileName := aliasFile + "." + uuid.New().String()

	marshal, err := yaml.Marshal(existingAliases)
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
	return existingAliases
}

// This function saves all the aliases for a specific resource.
func saveAliasesForResource(jsonApiType string, newAliases map[string]*AliasAttributes) {
	modifyAliases(jsonApiType, func(aliasMap map[string]*AliasAttributes) {

		// Aliases have the format KEY=VALUE and this maps to an ID.
		// This code checks for where two aliases have the same KEY and same ID, and replaces the old value, with the new one.
		// This happens in cases where we store a name like "name=John_Smith" and then the user renames it to "name=Jane_Doe".
		// The old alias for the same id name=John_Smith should be removed.
		for newAliasName, newAliasReferencedId := range newAliases {
			newAliasKeyName := strings.Split(newAliasName, "=")[0]
			for oldAliasName, oldAliasReferencedId := range aliasMap {
				oldAliasKeyName := strings.Split(oldAliasName, "=")[0]
				oldAliasValue := strings.Split(oldAliasName, "=")[1]

				if oldAliasKeyName == newAliasKeyName && oldAliasReferencedId.Id == newAliasReferencedId.Id {

					delete(aliasMap, oldAliasKeyName+"="+oldAliasValue)
				}
			}
		}

		for key, value := range newAliases {
			aliasMap[key] = value
		}
	})
}

func visitResources(data map[string]interface{}, prefix string, results map[string]map[string]*AliasAttributes) {
	if typeObj, typeKeyExists := data["type"]; typeKeyExists {
		if idObj, idKeyExists := data["id"]; idKeyExists {
			if typeKeyValue, typeKeyIsString := typeObj.(string); typeKeyIsString {
				if idKeyValue, idKeyIsString := idObj.(string); idKeyIsString {
					aliases := generateAliasesForStruct(typeKeyValue, idKeyValue, data)

					log.Tracef("Found a type and id pair %s => %s under prefix %s, aliases %v", typeKeyValue, idKeyValue, prefix, aliases)

					if _, ok := results[typeKeyValue]; !ok {
						results[typeKeyValue] = make(map[string]*AliasAttributes)
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

func generateAliasesForStruct(typeKey string, idKey string, data map[string]interface{}) map[string]*AliasAttributes {
	result := AliasAttributes{
		Id: idKey,
	}

	results := map[string]*AliasAttributes{
		// Identity, objects should be an alias of themselves.
		"id=" + idKey: &result,
	}

	jsonObjectsToInspect := make([]map[string]interface{}, 0)
	jsonObjectsToInspect = append(jsonObjectsToInspect, data)

	if val, ok := data["attributes"]; ok {
		if attributeVal, ok := val.(map[string]interface{}); ok {
			jsonObjectsToInspect = append(jsonObjectsToInspect, attributeVal)
		}
	}

	for _, jsonObjectToInspect := range jsonObjectsToInspect {

		if alias := getAliasForKey("name", jsonObjectToInspect); alias != "" {
			results[alias] = &result
		}

		if alias := getAliasForKey("sku", jsonObjectToInspect); alias != "" {
			results[alias] = &result
			result.Sku = getAttributeValueForKey("sku", jsonObjectToInspect)
		}

		if alias := getAliasForKey("slug", jsonObjectToInspect); alias != "" {
			results[alias] = &result
			result.Slug = getAttributeValueForKey("slug", jsonObjectToInspect)
		}

		if alias := getAliasForKey("email", jsonObjectToInspect); alias != "" {
			results[alias] = &result
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

func getAttributeValueForKey(key string, data map[string]interface{}) string {
	if val, ok := data[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		} else {
			return ""
		}
	} else {
		return ""
	}

}
