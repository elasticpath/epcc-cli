package variables

import (
	gojson "encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/elasticpath/epcc-cli/external/profiles"
	"github.com/google/uuid"
	"github.com/itchyny/gojq"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var mu sync.RWMutex
var vars = map[string]string{}
var loaded atomic.Bool
var dirty atomic.Bool

var directoryOverride = ""

func ensureLoaded() {
	if loaded.Load() {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	if loaded.Load() {
		return
	}

	data, err := os.ReadFile(getVariablesFile())
	if err != nil {
		log.Tracef("Could not read variables file: %v", err)
	} else {
		if err := yaml.Unmarshal(data, &vars); err != nil {
			log.Debugf("Could not unmarshal variables file: %v", err)
		} else {
			log.Tracef("Loaded %d variables from disk", len(vars))
		}
	}

	loaded.Store(true)
}

func GetVariable(name string) (string, bool) {
	ensureLoaded()
	mu.RLock()
	defer mu.RUnlock()
	v, ok := vars[name]
	return v, ok
}

// GetAllVariables returns a copy of all stored variables.
func GetAllVariables() map[string]string {
	ensureLoaded()
	mu.RLock()
	defer mu.RUnlock()
	result := make(map[string]string, len(vars))
	for k, v := range vars {
		result[k] = v
	}
	return result
}

func SetVariable(name, value string) {
	ensureLoaded()
	mu.Lock()
	defer mu.Unlock()
	vars[name] = value
	dirty.Store(true)
}

// ResolveVariableOrReturnIdentity resolves a "var/<name>" reference to its stored value.
// If the value does not have the "var/" prefix, it is returned as-is.
func ResolveVariableOrReturnIdentity(value string) string {
	if !strings.HasPrefix(value, "var/") {
		return value
	}
	name := value[4:]
	if v, ok := GetVariable(name); ok {
		return v
	}
	log.Warnf("Variable [%s] not found", name)
	return value
}

// ExtractAndSetVariables parses specs like "name=.jq.path" and extracts values from responseBody.
func ExtractAndSetVariables(specs []string, responseBody string) {
	for _, spec := range specs {
		parts := strings.SplitN(spec, "=", 2)
		if len(parts) != 2 {
			log.Warnf("Invalid --set-var spec %q, expected name=.jq.path", spec)
			continue
		}
		name := parts[0]
		jqPath := parts[1]

		results, err := runJQ(jqPath, responseBody)
		if err != nil {
			log.Warnf("Error running jq expression %q for variable %q: %v", jqPath, name, err)
			continue
		}

		if len(results) == 0 {
			log.Warnf("jq expression %q returned no results for variable %q, skipping", jqPath, name)
			continue
		}

		value := convertToString(results[0], name, jqPath)
		if value != nil {
			SetVariable(name, *value)
			log.Debugf("Set variable %q = %q", name, *value)
		}
	}
}

// runJQ executes a jq query on a JSON string and returns all results.
func runJQ(queryStr string, jsonStr string) ([]interface{}, error) {
	var obj interface{}
	if err := gojson.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return nil, err
	}

	query, err := gojq.Parse(queryStr)
	if err != nil {
		return nil, fmt.Errorf("error parsing jq expression %s: %w", queryStr, err)
	}

	iter := query.Run(obj)
	var results []interface{}
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, fmt.Errorf("error running jq expression %s: %w", queryStr, err)
		}
		results = append(results, v)
	}
	return results, nil
}

func convertToString(v interface{}, name string, jqPath string) *string {
	switch val := v.(type) {
	case string:
		return &val
	case nil:
		log.Warnf("jq expression %q returned null for variable %q, storing empty string", jqPath, name)
		empty := ""
		return &empty
	case float64:
		s := fmt.Sprintf("%v", val)
		return &s
	case bool:
		s := fmt.Sprintf("%v", val)
		return &s
	default:
		// object or array — marshal to JSON
		b, err := gojson.Marshal(val)
		if err != nil {
			log.Warnf("Could not marshal jq result for variable %q: %v", name, err)
			return nil
		}
		s := string(b)
		return &s
	}
}

func FlushVariables() {
	if !dirty.Load() {
		return
	}
	mu.RLock()
	defer mu.RUnlock()

	varFile := getVariablesFile()

	data, err := yaml.Marshal(vars)
	if err != nil {
		log.Warnf("Could not marshal variables: %v", err)
		return
	}

	dir := filepath.Dir(varFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		log.Warnf("Could not create variables directory: %v", err)
		return
	}

	tmpFile := varFile + "." + uuid.New().String()
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		log.Warnf("Could not write variables to temp file: %v", err)
		return
	}

	if err := os.Rename(tmpFile, varFile); err != nil {
		log.Warnf("Could not rename variables temp file: %v", err)
		return
	}

	dirty.Store(false)
	log.Debugf("Flushed %d variables to disk", len(vars))
}

func ClearAllVariables() {
	mu.Lock()
	defer mu.Unlock()
	vars = map[string]string{}
	dirty.Store(true)
	loaded.Store(true)
}

func getVariablesDirectory() string {
	if directoryOverride != "" {
		return directoryOverride
	}
	return filepath.FromSlash(profiles.GetProfileDataDirectory() + "/")
}

func getVariablesFile() string {
	return filepath.Clean(getVariablesDirectory() + "variables.yml")
}

// InitializeDirectoryForTesting sets a temporary directory for variable storage in tests.
func InitializeDirectoryForTesting() {
	dir, err := os.MkdirTemp("", "epcc-cli-variables-testing")
	if err != nil {
		log.Panic("Could not create directory", err)
	}
	directoryOverride = dir
	vars = map[string]string{}
	loaded.Store(true)
	dirty.Store(false)
}

// ClearTestState resets internal state for testing.
func ClearTestState() {
	mu.Lock()
	defer mu.Unlock()
	vars = map[string]string{}
	loaded.Store(false)
	dirty.Store(false)
	directoryOverride = ""
}
