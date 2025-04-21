package json

import (
	gojson "encoding/json"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/elasticpath/epcc-cli/external/templates"
	"github.com/itchyny/gojq"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

var segmentRegex = regexp.MustCompile("(.+?)(\\[[0-9]+])?$")

var attributeWithArrayIndex = regexp.MustCompile("\\[[0-9]+]")

func ToJson(args []string, noWrapping bool, compliant bool, attributes map[string]*resources.CrudEntityAttribute, useAliases bool) (string, error) {

	if len(args)%2 == 1 {
		return "", fmt.Errorf("the number of arguments %d supplied isn't even, json should be passed in key value pairs. Do you have an extra/missing id?", len(args))
	}

	firstArrayKeyIdx := -1
	firstFieldKeyIdx := -1
	for i := 0; i < len(args); i += 2 {
		key := args[i]
		if strings.HasPrefix(key, "[") {
			firstArrayKeyIdx = i
		} else {
			firstFieldKeyIdx = i
		}
	}

	if firstArrayKeyIdx >= 0 && firstFieldKeyIdx >= 0 {
		return "", fmt.Errorf("detected both array syntax arguments '%s' and object syntax arguments '%s'. Only one format can be used", args[firstArrayKeyIdx], args[firstFieldKeyIdx])
	}

	if firstArrayKeyIdx >= 0 {
		return toJsonArray(args, noWrapping, compliant, attributes, useAliases)
	} else {
		return toJsonObject(args, noWrapping, compliant, attributes, useAliases)
	}
}

func toJsonObject(args []string, noWrapping bool, compliant bool, attributes map[string]*resources.CrudEntityAttribute, useAliases bool) (string, error) {

	var result interface{} = make(map[string]interface{})
	var err error

	// 1. Collect all keys/values and group by prefix (e.g., data[0])
	grouped := make(map[string]map[string]string)
	for i := 0; i < len(args); i += 2 {
		key := args[i]
		val := args[i+1]
		prefix := key
		if idx := strings.Index(key, "."); idx != -1 {
			prefix = key[:idx]
		}
		if _, ok := grouped[prefix]; !ok {
			grouped[prefix] = make(map[string]string)
		}
		field := key
		if idx := strings.Index(key, "."); idx != -1 {
			field = key[idx+1:]
		}
		grouped[prefix][field] = val
	}

	// 2. Find all possible CONST attributes for data[n].*
	constFields := map[string]*resources.CrudEntityAttribute{}
	for attrKey, attr := range attributes {
		if strings.Contains(attrKey, "[n].") && strings.HasPrefix(attr.Type, "CONST:") {
			if idx := strings.Index(attrKey, "."); idx != -1 {
				field := attrKey[idx+1:]
				constFields[field] = attr
			}
		}
	}

	// 3. For each group, add missing const fields
	for _, fields := range grouped {
		for field, attr := range constFields {
			if _, ok := fields[field]; !ok {
				fields[field] = strings.TrimPrefix(attr.Type, "CONST:")
			}
		}
	}

	// 4. Now flatten back to args for the rest of the logic
	flatArgs := make([]string, 0, len(args)+(len(constFields)*len(grouped)))
	for prefix, fields := range grouped {
		for field, val := range fields {
			if field == prefix {
				flatArgs = append(flatArgs, prefix, val)
			} else {
				flatArgs = append(flatArgs, prefix+"."+field, val)
			}
		}
	}

	// 5. Continue with the original logic, but using flatArgs
	for i := 0; i < len(flatArgs); i += 2 {
		key := flatArgs[i]
		val := flatArgs[i+1]

		val = templates.Render(val)
		jsonKey := key
		switch {
		case key == "type" || key == "id":
			// These should always be in the root json object
		case strings.HasPrefix(key, "attributes.") || strings.HasPrefix(key, "relationships."):
			// We won't double encode these.
		case compliant:
			jsonKey = fmt.Sprintf("attributes.%s", key)
		default:
		}

		attributeName := key
		if strings.HasPrefix(key, "attributes.") {
			attributeName = strings.Replace(key, "attributes.", "", 1)
		}
		attributeName = attributeWithArrayIndex.ReplaceAllString(attributeName, "[n]")
		attributeInfo, hasAttribute := attributes[attributeName]
		useAttribute := false
		if hasAttribute {
			if strings.HasPrefix(attributeInfo.Type, "RESOURCE_ID:") {
				useAttribute = true
			}
			if strings.HasPrefix(attributeInfo.Type, "RESOURCE_ID:*") {
				useAttribute = false
			}
		}
		if useAttribute {
			if strings.HasPrefix(attributeInfo.Type, "RESOURCE_ID:") {
				resourceType := strings.Replace(attributeInfo.Type, "RESOURCE_ID:", "", 1)
				aliasAttributeToUse := "id"
				if attributeInfo.AliasAttribute != "" {
					aliasAttributeToUse = attributeInfo.AliasAttribute
				}
				if aliasType, ok := resources.GetResourceByName(resourceType); ok {
					if useAliases {
						val = aliases.ResolveAliasValuesOrReturnIdentity(aliasType.JsonApiType, aliasType.AlternateJsonApiTypesForAliases, val, aliasAttributeToUse)
					}
				} else {
					log.Warnf("Could not find a resource for %s, this is a bug.", resourceType)
				}
			}
		} else {
			splitAlias := strings.Split(val, "/")
			if len(splitAlias) == 4 {
				if splitAlias[0] == "alias" {
					if useAliases {
						val = aliases.ResolveAliasValuesOrReturnIdentity(splitAlias[1], []string{}, splitAlias[2], splitAlias[3])
					}
				}
			}
		}
		val = formatValue(val)
		arrayNotationPath := ""
		for _, str := range strings.Split(jsonKey, ".") {
			arrayNotationPath += segmentRegex.ReplaceAllString(str, "[\"$1\"]$2")
		}
		query := fmt.Sprintf(".%s=%s", arrayNotationPath, val)
		result, err = RunJQ(query, result)
		if err != nil {
			return "{}", err
		}
	}

	if !noWrapping {
		result, err = RunJQ(`{ "data": . }`, result)
	}
	jsonStr, err := gojson.Marshal(result)
	return string(jsonStr), err
}

func toJsonArray(args []string, noWrapping bool, compliant bool, attributes map[string]*resources.CrudEntityAttribute, useAliases bool) (string, error) {

	var result interface{} = make([]interface{}, 0)
	var err error

	// Group all keys by array index (e.g. data[0], data[1])
	grouped := make(map[string]map[string]string)
	for i := 0; i < len(args); i += 2 {
		key := args[i]
		val := args[i+1]
		// Extract prefix up to array index, e.g. data[0]
		prefix := key
		if idx := strings.Index(key, "."); idx != -1 {
			prefix = key[:idx]
		}
		if _, ok := grouped[prefix]; !ok {
			grouped[prefix] = make(map[string]string)
		}
		field := key
		if idx := strings.Index(key, "."); idx != -1 {
			field = key[idx+1:]
		}
		grouped[prefix][field] = val
	}

	// Find all possible CONST attributes for data[n].*
	constFields := map[string]*resources.CrudEntityAttribute{}
	for attrKey, attr := range attributes {
		if strings.Contains(attrKey, "[n].") && strings.HasPrefix(attr.Type, "CONST:") {
			// attrKey example: data[n].type
			if idx := strings.Index(attrKey, "."); idx != -1 {
				field := attrKey[idx+1:]
				constFields[field] = attr
			}
		}
	}

	// For each array element, add consts if not present
	arrayResult := make([]map[string]interface{}, 0, len(grouped))
	for _, fields := range grouped {
		obj := map[string]interface{}{}
		for k, v := range fields {
			obj[k] = formatValue(v)
		}
		// Add missing const fields
		for field, attr := range constFields {
			if _, ok := obj[field]; !ok {
				constVal := strings.TrimPrefix(attr.Type, "CONST:")
				obj[field] = constVal
			}
		}
		arrayResult = append(arrayResult, obj)
	}

	result = arrayResult
	if !noWrapping {
		result = map[string]interface{}{"data": result}
	}
	jsonStr, err := gojson.Marshal(result)
	return string(jsonStr), err
}

func RunJQOnString(queryStr string, json string) (interface{}, error) {

	var obj interface{}

	err := gojson.Unmarshal([]byte(json), &obj)

	if err != nil {
		return nil, err
	}

	return RunJQ(queryStr, obj)
}

func RunJQOnStringAndGetString(queryStr string, json string) (string, error) {
	result, err := RunJQOnString(queryStr, json)

	if err != nil {
		return "", err
	}

	if result, ok := result.(string); ok {
		return result, nil
	}

	return "", fmt.Errorf("could not convert %T into string", result)
}

func RunJQOnStringAndMarshalResponse(queryStr string, json string, obj any) error {
	result, err := RunJQOnString(queryStr, json)

	if err != nil {
		return err
	}

	return mapstructure.Decode(result, &obj)
}

func RunJQ(queryStr string, result interface{}) (interface{}, error) {
	query, err := gojq.Parse(queryStr)

	if err != nil {
		// %w causes the error to be wrapped.
		return nil, fmt.Errorf("error parsing json key %s: %w", queryStr, err)
	}

	iter := query.Run(result)

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			partialResult, _ := gojson.Marshal(result)

			return nil, fmt.Errorf("error %w when running query %s on json %s", err, queryStr, partialResult)
		}

		result = v
	}
	return result, nil
}

// These copy and paste functions below, were because the existing functions were buggy
// if the iterator returns more than one thing, only the last is saved.
// it was deemed to risky to fix at the moment.
func RunJQOnStringWithArray(queryStr string, json string) ([]interface{}, error) {

	var obj interface{}

	err := gojson.Unmarshal([]byte(json), &obj)

	if err != nil {
		return nil, err
	}

	return RunJQWithArray(queryStr, obj)
}

func RunJQWithArray(queryStr string, result interface{}) ([]interface{}, error) {
	query, err := gojq.Parse(queryStr)

	if err != nil {
		// %w causes the error to be wrapped.
		return nil, fmt.Errorf("error parsing json key %s: %w", queryStr, err)
	}

	iter := query.Run(result)

	queryResult := []interface{}{}

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			partialResult, _ := gojson.Marshal(result)

			return nil, fmt.Errorf("error %w when running query %s on json %s", err, queryStr, partialResult)
		}

		queryResult = append(queryResult, v)
	}
	return queryResult, nil
}

func formatValue(v string) string {
	if match, _ := regexp.MatchString("^(-?[0-9]+(\\.[0-9]+)?|false|true|null)$", v); match {
		return v
	} else if match, _ := regexp.MatchString("^\\\".+\\\"$", v); match {
		return v
	} else if match, _ := regexp.MatchString("^\\[\\]$", v); match {
		return v
	} else {
		v = strings.ReplaceAll(v, "\\", "\\\\")
		v = strings.ReplaceAll(v, `"`, `\"`)

		return fmt.Sprintf("\"%s\"", v)
	}
}
