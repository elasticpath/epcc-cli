package autofill

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/elasticpath/epcc-cli/external/resources"
	log "github.com/sirupsen/logrus"
)

var faker = gofakeit.New(0)

var zeroValue = reflect.Value{}

func GetAutoFillQueryParameters(resourceName string, qps []resources.QueryParameter) []string {
	args := make([]string, 0)

	for _, v := range qps {
		key := v.Name
		autoFill := v.AutoFill
		args = processAutoFill(resourceName, autoFill, args, key, key)
	}

	return args
}
func GetJsonArrayForResource(r *resources.Resource) []string {

	args := make([]string, 0)

	for attributeName, data := range r.Attributes {
		key := data.Key
		key = strings.Replace(key, "data[n]", "data[0]", 1)
		autofill := data.AutoFill

		args = processAutoFill(r.SingularName, autofill, args, key, attributeName)

	}
	return args

}

func processAutoFill(resourceName string, autofill string, args []string, key string, attributeName string) []string {
	if strings.HasPrefix(autofill, "FUNC:") {

		v := reflect.ValueOf(faker)
		methodName := strings.Trim(autofill[5:], " ")

		var rejectZero = false
		if strings.HasPrefix(methodName, "NonZero") {
			methodName = methodName[7:]
			rejectZero = true
		}

		method := v.MethodByName(methodName)
		if method.IsValid() {
			result := method.Call([]reflect.Value{})
			if len(result) == 1 {

				var arg string

				switch {
				case result[0].CanUint():
					for true {
						v := result[0].Uint()
						if v == 0 && rejectZero {
							result = method.Call([]reflect.Value{})
							continue
						}
						arg = strconv.FormatUint(v, 10)
						break
					}

				case result[0].CanInt():
					for true {
						v := result[0].Int()
						if v == 0 && rejectZero {
							result = method.Call([]reflect.Value{})
							continue
						}
						arg = strconv.FormatInt(v, 10)
						break
					}
				default:
					arg = result[0].String()

					if _, err := strconv.Atoi(arg); err == nil {
						// If we get an integer value back, lets just quote it.
						arg = fmt.Sprintf("\"%s\"", arg)
					}
				}

				args = append(args, key, arg)
			} else {
				log.Warnf("Got unexpected number of results from calling %s -> %d", methodName, len(result))
			}

		} else {
			log.Warnf("Could not find autofill method %s for attribute %s on resource %s", methodName, attributeName, resourceName)
		}

	} else if strings.HasPrefix(autofill, "VALUE:") {
		args = append(args, key, strings.Trim(autofill[6:], " "))
	}
	return args
}
