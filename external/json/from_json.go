package json

import (
	gojson "encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/yukithm/json2csv"
	"github.com/yukithm/json2csv/jsonpointer"
)

// FromJson is the inverse operation of ToJson it converts a json object into the key value pairs we would type on the command line
func FromJson(json string) ([]string, error) {

	o, err := FromJsonToMap(json)

	if err != nil {
		return []string{}, err
	}

	sortedKeys := make([]string, 0, len(o))
	for k := range o {
		sortedKeys = append(sortedKeys, k)
	}

	sort.Strings(sortedKeys)

	out := make([]string, 0, len(sortedKeys))
	for _, k := range sortedKeys {
		out = append(out, k, o[k])
	}

	return out, nil
}

// FromJsonToMap converts a json string into a map of key value pairs.
func FromJsonToMap(json string) (map[string]string, error) {
	var obj any

	out := map[string]string{}

	err := gojson.Unmarshal([]byte(json), &obj)

	if err != nil {
		return nil, err
	}

	kvs, err := json2csv.JSON2CSV(obj)
	if err != nil {

		return nil, fmt.Errorf("error during processing (csv): %v", err)

	}

	if len(kvs) == 0 {
		return out, nil
	}

	if len(kvs) != 1 {
		return nil, fmt.Errorf("more than one result came back")
	}

	if len(kvs) != 1 {
		return nil, fmt.Errorf("more than one result came back")
	}

	for _, kv := range kvs {

		keys := kv.Keys()

		sort.Strings(keys)

		for _, k := range keys {
			v := kv[k]

			jp, err := jsonpointer.New(k)

			if err != nil {
				return nil, fmt.Errorf("error during processing (jp): %v", err)
			}

			outKey := jp.DotNotation(true)

			// We need to check first for prefixes, because a key like data.attributes.data should end up like data
			if strings.HasPrefix(outKey, "data.attributes.") {
				outKey = strings.TrimPrefix(outKey, "data.attributes.")
			} else if strings.HasPrefix(outKey, "data.") {
				outKey = strings.TrimPrefix(outKey, "data.")
			}

			if s, ok := v.(string); ok {
				out[outKey] = fmt.Sprintf("\"%s\"", s)
			} else if s, ok := v.(float64); ok {
				if s == float64(int(s)) {
					out[outKey] = fmt.Sprintf("%d", int(s))
				} else {
					out[outKey] = fmt.Sprintf("%g", s)
				}
			} else if s, ok := v.(bool); ok {
				out[outKey] = fmt.Sprintf("%t", s)
			} else {
				return nil, fmt.Errorf("error during processing (jp value), unknown type (%T) for %v", v, v)
			}
		}

	}

	return out, nil
}
