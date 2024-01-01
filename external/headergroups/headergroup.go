package headergroups

import (
	"encoding/json"
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/profiles"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

var headerGroups = map[string]map[string]string{}

var headerGroupMutex = sync.RWMutex{}

var headerGroupsLoaded = atomic.Bool{}

var headerToAliasType = map[string]string{}

var headerToAliasMutex = sync.RWMutex{}

func AddHeaderAliasMapping(header string, resource string) {

	headerToAliasMutex.Lock()
	defer headerToAliasMutex.Unlock()

	headerToAliasType[strings.ToLower(header)] = resource

}

func ClearAllHeaderAliasMappings() {
	headerToAliasMutex.Lock()
	defer headerToAliasMutex.Unlock()

	headerToAliasType = map[string]string{}
}

func AddHeaderGroup(name string, headers map[string]string) {
	initializeHeaderGroups()
	headerGroupMutex.Lock()
	defer headerGroupMutex.Unlock()

	headerToAliasMutex.RLock()
	defer headerToAliasMutex.RUnlock()

	headerGroups[name] = map[string]string{}

	for k, v := range headers {
		aliasType := headerToAliasType[strings.ToLower(k)]

		if aliasType != "" {
			headerGroups[name][k] = aliases.ResolveAliasValuesOrReturnIdentity(aliasType, []string{}, v, "id")
		} else {
			headerGroups[name][k] = v
		}

	}
}

func AddHeaderToGroup(name string, header string, value string) {
	initializeHeaderGroups()
	headerGroupMutex.Lock()
	defer headerGroupMutex.Unlock()

	headerToAliasMutex.RLock()
	defer headerToAliasMutex.RUnlock()

	if headerGroups[name] == nil {
		headerGroups[name] = map[string]string{}
	}

	aliasType := headerToAliasType[strings.ToLower(header)]

	if aliasType != "" {
		headerGroups[name][header] = aliases.ResolveAliasValuesOrReturnIdentity(aliasType, []string{}, value, "id")
	} else {
		headerGroups[name][header] = value
	}
}

func RemoveHeaderGroup(name string) {
	initializeHeaderGroups()
	headerGroupMutex.Lock()
	defer headerGroupMutex.Unlock()

	delete(headerGroups, name)

}

func RemoveHeaderFromGroup(name string, header string) {
	initializeHeaderGroups()
	headerGroupMutex.Lock()
	defer headerGroupMutex.Unlock()

	if headerGroups[name] == nil {
		return
	}

	delete(headerGroups[name], header)

	if len(headerGroups[name]) == 0 {
		delete(headerGroups, name)
	}
}

func GetAllHeaders() map[string]string {
	initializeHeaderGroups()
	headerGroupMutex.RLock()
	defer headerGroupMutex.RUnlock()

	headers := map[string]string{}

	for s, headerGroup := range headerGroups {
		for k, v := range headerGroup {
			if headers[k] != "" {
				log.Warnf("Duplicate header found in group %s, overwriting: %s", s, k)
			}
			headers[k] = v
		}
	}

	return headers
}

func GetAllHeaderGroups() []string {
	initializeHeaderGroups()
	headerGroupMutex.RLock()
	defer headerGroupMutex.RUnlock()

	groups := []string{}

	for s := range headerGroups {
		groups = append(groups, s)
	}

	return groups
}

func ClearAllHeaderGroups() {
	initializeHeaderGroups()
	headerGroupMutex.Lock()
	defer headerGroupMutex.Unlock()

	headerGroups = map[string]map[string]string{}

	path := GetHeaderGroupPath()
	err := os.Remove(path)

	if err == nil || os.IsNotExist(err) {
		return
	}

	log.Warnf("Could not delete header groups(%s): %v", path, err)

}

func initializeHeaderGroups() {
	if headerGroupsLoaded.Load() {
		return
	} else {
		headerGroupMutex.Lock()
		defer headerGroupMutex.Unlock()
		loadHeaderGroupsFromDisk()
	}
}

func GetHeaderGroupPath() string {
	return filepath.Clean(profiles.GetProfileDataDirectory() + "/header_groups.json")
}

func loadHeaderGroupsFromDisk() {
	headerGroupPath := GetHeaderGroupPath()
	data, err := os.ReadFile(headerGroupPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Warnf("Could not read %s, error %s", headerGroupPath, err)
		} else {
			log.Debugf("Error occurred while reading header group path %s: %v", headerGroupPath, err)
		}
		data = []byte{}
	} else {

		err = json.Unmarshal(data, &headerGroups)
		if err != nil {
			log.Debugf("Could not unmarshall existing file %s, error %s", data, err)
		}
	}
	log.Tracef("Retrieved %s from disk", headerGroups)
	headerGroupsLoaded.Store(true)
}

func saveHeaderGroupsToDisk() {
	jsonHeaderGroups, err := json.Marshal(headerGroups)

	headerGroupPath := GetHeaderGroupPath()
	log.Debugf("Saving header groups to disk in %v", headerGroupPath)
	log.Tracef("Saving all data %s", jsonHeaderGroups)
	if err != nil {
		log.Warnf("Could not convert token to JSON  %v", err)
	} else {
		err = os.WriteFile(headerGroupPath, jsonHeaderGroups, 0600)

		if err != nil {
			log.Warnf("Could not save token %s, error: %v", headerGroupPath, err)
		} else {
			log.Debugf("Saved token to %s", headerGroupPath)
		}
	}

}

func SyncHeaderGroups() {
	initializeHeaderGroups()

	headerGroupMutex.RLock()
	defer headerGroupMutex.RUnlock()

	saveHeaderGroupsToDisk()
}

func FlushHeaderGroups() {

	initializeHeaderGroups()
	headerGroupMutex.Lock()
	defer headerGroupMutex.Unlock()

	saveHeaderGroupsToDisk()
	headerGroupsLoaded.Store(false)
	headerGroups = map[string]map[string]string{}

}
