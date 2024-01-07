package json

import (
	gojson "encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
)

var KeysToDelete = []string{
	".data.links",
	".data.meta",
	".data[].links",
	".data[].meta",
	".links",
	".meta",
	"..|nulls",
	`..|select(. == {})`,
	`..|select(. == {})`,
	`..|select(. == "")`,
}

func Compact(json string) (string, error) {

	var obj interface{}

	err := gojson.Unmarshal([]byte(json), &obj)

	if err != nil {
		return "", err
	}

	for _, key := range KeysToDelete {
		newObj, err := RunJQWithArray(fmt.Sprintf("del(%s)", key), obj)

		if err == nil {
			if len(newObj) == 1 {
				obj = newObj[0]
			} else {
				log.Warnf("Couldn't compact with key %s, due to unexpected result size", key)
				return json, nil
			}
		}
	}

	str, err := gojson.Marshal(obj)

	if err != nil {
		return "", err
	}

	return string(str), err
}
