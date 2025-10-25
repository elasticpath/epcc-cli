package autofill

import (
	"strings"

	"github.com/elasticpath/epcc-cli/external/faker"
	"github.com/elasticpath/epcc-cli/external/resources"
)

func GetAutoFillQueryParameters(resourceName string, qps []resources.QueryParameter) []string {
	args := make([]string, 0)

	for _, v := range qps {
		key := v.Name
		autoFill := v.AutoFill
		args = processAutoFill(autoFill, args, key)
	}

	return args
}

func GetJsonArrayForResource(r *resources.Resource) []string {

	args := make([]string, 0)

	for _, data := range r.Attributes {
		key := data.Key
		key = strings.Replace(key, "data[n]", "data[0]", 1)
		autofill := data.AutoFill

		args = processAutoFill(autofill, args, key)

	}
	return args

}

func processAutoFill(autofill string, args []string, key string) []string {
	if strings.HasPrefix(autofill, "FUNC:") {

		methodName := strings.Trim(autofill[5:], " ")

		v := faker.CallFakeFunc(methodName)
		args = append(args, key, v)

	} else if strings.HasPrefix(autofill, "VALUE:") {
		args = append(args, key, strings.Trim(autofill[6:], " "))
	}
	return args
}
