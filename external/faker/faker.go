package faker

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/brianvoe/gofakeit/v6"
	log "github.com/sirupsen/logrus"
)

var faker = gofakeit.New(1)

func Seed(n int64) {
	faker = gofakeit.New(n)
}

func CallFakeFunc(methodName string) string {

	var arg string
	v := reflect.ValueOf(faker)

	var rejectZero = false
	if strings.HasPrefix(methodName, "NonZero") {
		methodName = methodName[7:]
		rejectZero = true
	}

	method := v.MethodByName(methodName)
	if method.IsValid() {
		result := method.Call([]reflect.Value{})
		if len(result) == 1 {

			switch {
			case result[0].CanUint():
				for {
					v := result[0].Uint()
					if v == 0 && rejectZero {
						result = method.Call([]reflect.Value{})
						continue
					}
					arg = strconv.FormatUint(v, 10)
					break
				}

			case result[0].CanInt():
				for {
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

		} else {
			log.Warnf("Got unexpected number of results from calling %s -> %d", methodName, len(result))
		}

	} else {
		log.Warnf("Could not find autofill method %s", methodName)
	}

	return arg
}
