package autofill

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/elasticpath/epcc-cli/external/resources"
	log "github.com/sirupsen/logrus"
	"reflect"
	"strconv"
	"strings"
)

var faker = gofakeit.New(0)

var zeroValue = reflect.Value{}

func GetJsonArrayForResource(r *resources.Resource) []string {

	args := make([]string, 0)

	for attributeName, data := range r.Attributes {
		autofill := data.AutoFill

		if strings.HasPrefix(autofill, "FUNC:") {

			v := reflect.ValueOf(faker)
			methodName := strings.Trim(autofill[5:], " ")
			method := v.MethodByName(methodName)
			if method.IsValid() {
				result := method.Call([]reflect.Value{})
				if len(result) == 1 {

					arg := result[0].String()

					args = append(args, data.Key, arg)
				} else {
					log.Warnf("Got unexpected number of results from calling %s -> %d", methodName, len(result))
				}

			} else {
				log.Warnf("Could not find autofill method %s for attribute %s on resource %s", methodName, attributeName, r.SingularName)
			}

		} else if strings.HasPrefix(autofill, "VALUE:") {
			args = append(args, data.Key, strings.Trim(autofill[6:], " "))
		}

	}

	for k, v := range args {
		if _, err := strconv.Atoi(v); err == nil {
			// If we get an integer value back, lets just quote it.
			args[k] = fmt.Sprintf("\"%s\"", v)
		}
	}

	return args

}
