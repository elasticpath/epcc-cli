package json

import (
	gojson "encoding/json"
	"fmt"
	"github.com/itchyny/gojq"
	"log"
	"regexp"
	"strings"
)

var segmentRegex = regexp.MustCompile("(.+?)(\\[[0-9]+])?$")

func ToJson(args []string, noWrapping bool, compliant bool) (string, error) {

	if len(args)%2 == 1 {
		return "", fmt.Errorf("The number arguments %d supplied isn't even, json should be passed in key value pairs", len(args))
	}

	result := make(map[string]interface{})

	var err error

	for i := 0; i < len(args); i += 2 {
		key := args[i]
		val := formatValue(args[i+1])

		jsonKey := key
		switch {
		case key == "type" || key == "id":
			// These should always be in the root json object
		case strings.HasPrefix("attributes.", key) || strings.HasPrefix("relationships.", key):
			// We won't double encode these.
		case compliant:
			jsonKey = fmt.Sprintf("attributes.%s", key)
		default:

		}
		arrayNotationPath := ""

		for _, str := range strings.Split(jsonKey, ".") {
			arrayNotationPath += segmentRegex.ReplaceAllString(str, "[\"$1\"]$2")
		}

		query := fmt.Sprintf(".%s=%s", arrayNotationPath, val)

		result, err = runJQ(query, result)
		if err != nil {
			return "{}", err
		}

	}

	if !noWrapping {
		result, err = runJQ(`{ "data": . }`, result)
	}

	jsonStr, err := gojson.Marshal(result)

	return string(jsonStr), err

}

func runJQ(queryStr string, result map[string]interface{}) (map[string]interface{}, error) {
	query, err := gojq.Parse(queryStr)

	if err != nil {
		// %w causes the error to be wrapped.
		return nil, fmt.Errorf("Error parsing json key %s: %w", queryStr, err)
	}

	iter := query.Run(result)

	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			log.Fatalln(err)
		}
		if res2, ok := v.(map[string]interface{}); ok {
			result = res2
		}

	}
	return result, nil
}

func formatValue(v string) string {
	if match, _ := regexp.MatchString("^([0-9]+|false|true|null)$", v); match {
		return v
	} else if match, _ := regexp.MatchString("^\\\".+\\\"$", v); match {
		return v
	} else if match, _ := regexp.MatchString("^\\[\\]$", v); match {
		return v
	} else {
		return fmt.Sprintf("\"%s\"", v)
	}

}
