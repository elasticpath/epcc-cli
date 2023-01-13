package json

import (
	gojson "encoding/json"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/itchyny/gojq"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

var segmentRegex = regexp.MustCompile("(.+?)(\\[[0-9]+])?$")

var attributeWithArrayIndex = regexp.MustCompile("\\[[0-9]+]")

func ToJson(args []string, noWrapping bool, compliant bool, attributes map[string]*resources.CrudEntityAttribute) (string, error) {

	if len(args)%2 == 1 {
		return "", fmt.Errorf("the number arguments %d supplied isn't even, json should be passed in key value pairs", len(args))
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
		return toJsonArray(args, noWrapping, compliant, attributes)
	} else {
		return toJsonObject(args, noWrapping, compliant, attributes)
	}
}

func toJsonObject(args []string, noWrapping bool, compliant bool, attributes map[string]*resources.CrudEntityAttribute) (string, error) {

	var result interface{} = make(map[string]interface{})

	var err error

	for i := 0; i < len(args); i += 2 {
		key := args[i]
		val := args[i+1]

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

		// Look for a match of attribute names with [n] instead of [0] or [1] whatever the user supplied
		attributeName = attributeWithArrayIndex.ReplaceAllString(attributeName, "[n]")

		if attributeInfo, ok := attributes[attributeName]; ok {
			if strings.HasPrefix(attributeInfo.Type, "RESOURCE_ID:") {
				resourceType := strings.Replace(attributeInfo.Type, "RESOURCE_ID:", "", 1)

				aliasAttributeToUse := "id"
				if attributeInfo.AliasAttribute != "" {
					aliasAttributeToUse = attributeInfo.AliasAttribute
				}
				if aliasType, ok := resources.GetResourceByName(resourceType); ok {
					val = aliases.ResolveAliasValuesOrReturnIdentity(aliasType.JsonApiType, aliasType.AlternateJsonApiTypesForAliases, val, aliasAttributeToUse)
				} else {
					log.Warnf("Could not find a resource for %s, this is a bug.", resourceType)
				}
			}
		} else {
			splitAlias := strings.Split(val, "/")

			if len(splitAlias) == 4 {
				if splitAlias[0] == "alias" {
					val = aliases.ResolveAliasValuesOrReturnIdentity(splitAlias[1], []string{}, splitAlias[2], splitAlias[3])
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

func toJsonArray(args []string, noWrapping bool, compliant bool, attributes map[string]*resources.CrudEntityAttribute) (string, error) {

	var result interface{} = make([]interface{}, 0)

	var err error

	for i := 0; i < len(args); i += 2 {
		key := args[i]
		val := args[i+1]

		jsonKey := key

		val = formatValue(val)

		query := fmt.Sprintf(".%s |= %s", jsonKey, val)

		result, err = RunJQ(query, result)
		if err != nil {
			return "[]", err
		}

	}

	if !noWrapping {
		result, err = RunJQ(`{ "data": . }`, result)
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
	if match, _ := regexp.MatchString("^([0-9]+(\\.[0-9]+)?|false|true|null)$", v); match {
		return v
	} else if match, _ := regexp.MatchString("^\\\".+\\\"$", v); match {
		return v
	} else if match, _ := regexp.MatchString("^\\[\\]$", v); match {
		return v
	} else {
		return fmt.Sprintf("\"%s\"", v)
	}

}
