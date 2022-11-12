package aliases

import (
	"encoding/json"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/id"
	"github.com/elasticpath/epcc-cli/external/profiles"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var mutex = &sync.RWMutex{}

var aliasDirectoryOverride = ""
var aliases = map[string]map[string]*id.IdableAttributes{}

var dirtyAliases = map[string]bool{}

var SkipAliasProcessing = false

func ClearAllAliasesForJsonApiType(jsonApiType string) error {
	ClearCache(jsonApiType)

	if err := os.Remove(getAliasFileForJsonApiType(getAliasDataDirectory(), jsonApiType)); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// Used to determine index of element in array.
var arrayPathPattern = regexp.MustCompile("^\\.data\\[([0-9]+)]$")

// Used to determine name of relationship
var relationshipPattern = regexp.MustCompile("^\\.data(?:\\[[0-9]+])?\\.relationships\\.([^.]+)\\.data")

func GetAliasesForJsonApiType(jsonApiType string) map[string]*id.IdableAttributes {

	mutex.RLock()

	aliasMap, ok := aliases[jsonApiType]

	if !ok {
		mutex.RUnlock()
		mutex.Lock()
		defer mutex.Unlock()
		aliasFile := getAliasFileForJsonApiType(getAliasDataDirectory(), jsonApiType)

		aliasMap = map[string]*id.IdableAttributes{}

		data, err := os.ReadFile(aliasFile)
		if err != nil {
			log.Debugf("Could not read %s, error %s", aliasFile, err)
			data = []byte{}
		} else {
		}

		err = yaml.Unmarshal(data, aliasMap)
		aliases[jsonApiType] = aliasMap
		if err != nil {
			log.Debugf("Could not unmarshall existing file %s, error %s", data, err)
		} else {
			log.Debugf("Aliases for type [%s] loaded, with %d aliases", jsonApiType, len(aliasMap))
		}

	} else {
		mutex.RUnlock()
	}

	return aliasMap
}

func ResolveAliasValuesOrReturnIdentity(jsonApiType string, value string, attribute string) string {
	if result, ok := GetAliasesForJsonApiType(jsonApiType)[value]; ok {

		if attribute == "id" {
			return result.Id
		}
		if attribute == "slug" {
			return result.Slug
		}

		if attribute == "sku" {
			return result.Sku
		}

		if attribute == "code" {
			return result.Code
		}

	}
	return value
}

func SaveAliasesForResources(jsonTxt string) {
	if SkipAliasProcessing {
		return
	}
	var jsonStruct = map[string]interface{}{}
	err := json.Unmarshal([]byte(jsonTxt), &jsonStruct)
	if err != nil {
		log.Warnf("Response was not JSON so not scanning for aliases")
		return
	}

	results := map[string]map[string]*id.IdableAttributes{}
	visitResources(jsonStruct, "", "", map[string]*id.IdableAttributes{}, results)

	log.Tracef("All aliases found in JSON: %v", results)

	for resourceType, foundAliases := range results {
		saveAliasesForResource(resourceType, foundAliases)
		log.Tracef("Number of resources for type [%s] is now %d and value is %v", resourceType, len(aliases[resourceType]), aliases[resourceType])
	}

}

func DeleteAliasesById(idStr string, jsonApiType string) {
	modifyAliases(jsonApiType, func(m map[string]*id.IdableAttributes) {
		for key, value := range m {
			if value.Id == idStr {
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

func modifyAliases(jsonApiType string, fn func(map[string]*id.IdableAttributes)) {
	aliasMap := GetAliasesForJsonApiType(jsonApiType)

	mutex.Lock()
	defer mutex.Unlock()

	fn(aliasMap)
	dirtyAliases[jsonApiType] = true

}

// This function saves all the aliases for a specific resource.
func saveAliasesForResource(jsonApiType string, newAliases map[string]*id.IdableAttributes) {

	modifyAliases(jsonApiType, func(aliasMap map[string]*id.IdableAttributes) {

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

func visitResources(data map[string]interface{}, prefix string, parentAliasType string, parentAliases map[string]*id.IdableAttributes, results map[string]map[string]*id.IdableAttributes) {
	aliases := map[string]*id.IdableAttributes{}
	myType := ""
	if typeObj, typeKeyExists := data["type"]; typeKeyExists {
		if idObj, idKeyExists := data["id"]; idKeyExists {
			if typeKeyValue, typeKeyIsString := typeObj.(string); typeKeyIsString {
				if idKeyValue, idKeyIsString := idObj.(string); idKeyIsString {

					myType = typeKeyValue
					aliases = generateAliasesForStruct(prefix, parentAliasType, parentAliases, typeKeyValue, idKeyValue, data)

					log.Tracef("Found a type and id pair %s => %s under prefix %s, aliases %v", typeKeyValue, idKeyValue, prefix, aliases)

					if _, ok := results[typeKeyValue]; !ok {
						results[typeKeyValue] = make(map[string]*id.IdableAttributes)
					}

					for aliasKey, aliasValue := range aliases {
						results[typeKeyValue][aliasKey] = aliasValue
					}
				}
			}
		}
	}

	parentAliasesToUse := parentAliases

	if len(aliases) > 0 {
		parentAliasesToUse = aliases
	}

	parentAliasTypeToUse := parentAliasType
	if myType != "" {
		parentAliasTypeToUse = myType
	}
	// Recursively descend over each element
	for key, val := range data {
		if mapType, ok := val.(map[string]interface{}); ok {
			visitResources(mapType, prefix+"."+key, parentAliasTypeToUse, parentAliasesToUse, results)
		}

		if arrayType, ok := val.([]interface{}); ok {
			for idx, value := range arrayType {
				if mapValue, ok := value.(map[string]interface{}); ok {
					visitResources(mapValue, prefix+"."+key+"["+strconv.Itoa(idx)+"]", parentAliasTypeToUse, parentAliasesToUse, results)
				}

			}
		}

	}

	return
}

func generateAliasesForStruct(prefix string, parentAliasType string, parentAliases map[string]*id.IdableAttributes, typeKey string, idKey string, data map[string]interface{}) map[string]*id.IdableAttributes {
	result := id.IdableAttributes{
		Id: idKey,
	}

	results := map[string]*id.IdableAttributes{
		// Identity, objects should be an alias of themselves.
		"id=" + idKey: &result,
	}

	if prefix == ".data" {
		results["last_read=entity"] = &result
	}

	if arrayPathPattern.MatchString(prefix) {
		matches := arrayPathPattern.FindStringSubmatch(prefix)
		results["last_read=array["+matches[1]+"]"] = &result
	}

	if relationshipPattern.MatchString(prefix) {
		matches := relationshipPattern.FindStringSubmatch(prefix)
		//related_buz_for_foo_id_123
		keyPrefix := "related_" + matches[1] + "_for_" + parentAliasType + "_"

		for k := range parentAliases {
			results[keyPrefix+k] = &result
		}

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

		if alias := getAliasForKey("code", jsonObjectToInspect); alias != "" {
			results[alias] = &result
			result.Code = getAttributeValueForKey("code", jsonObjectToInspect)
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

func InitializeAliasDirectoryForTesting() {
	dir, err := os.MkdirTemp("", "epcc-cli-aliases-testing")
	if err != nil {
		log.Panic("Could not create directory", err)
	}

	aliasDirectoryOverride = dir
	log.Infof("Alias directory for tests is %s", dir)
}

func ClearAllCaches() {
	mutex.Lock()
	aliases = map[string]map[string]*id.IdableAttributes{}
	dirtyAliases = map[string]bool{}
	mutex.Unlock()
}

func ClearCache(jsonApiType string) {
	mutex.Lock()
	delete(aliases, jsonApiType)
	delete(dirtyAliases, jsonApiType)
	mutex.Unlock()
}

func ClearAllAliases() error {
	aliasDataDirectory := getAliasDataDirectory()

	if err := os.RemoveAll(aliasDataDirectory); err != nil && !os.IsNotExist(err) {
		return err
	}

	ClearAllCaches()
	return nil

}

func FlushAliases() int {
	changed := SyncAliases()
	ClearAllCaches()

	return changed
}

func SyncAliases() int {

	syncedFiles := 0
	mutex.RLock()
	defer mutex.RUnlock()

	for jsonApiType, val := range dirtyAliases {
		if val == false {
			log.Errorf("Not expecting a dirty alias to be false, should either exist or not, this is a bug, for type %s", jsonApiType)
			continue
		}

		aliasFile := path.Clean(getAliasFileForJsonApiType(getAliasDataDirectory(), jsonApiType))

		aliasesForType := aliases[jsonApiType]

		// We will write to a temp file and then rename, to prevent data loss. rename's in the same folder are likely atomic in most settings.
		// Although we should probably sync on the file as well, that might be too much overhead, and I was too lazy to rewrite this
		// https://github.com/golang/go/issues/20599
		tmpFileName := aliasFile + "." + uuid.New().String()

		marshal, err := yaml.Marshal(aliasesForType)

		if err != nil {
			log.Warnf("Could not save aliases for %s, error %v", tmpFileName, err)
			continue
		}

		err = os.WriteFile(tmpFileName, marshal, 0600)
		if err != nil {
			log.Warnf("Could not save aliases for %s, error %v", tmpFileName, err)
			continue
		}

		err = os.Rename(tmpFileName, aliasFile)
		if err != nil {
			log.Warnf("Could not save aliases for %s, error %v", tmpFileName, err)
			continue
		}

		syncedFiles++
		delete(dirtyAliases, jsonApiType)
		log.Tracef("Successfully wrote aliases to disk for file %s in %s", jsonApiType, aliasFile)
	}

	log.Debugf("Syncing aliases to disk, %d files changed", syncedFiles)
	return syncedFiles

}
