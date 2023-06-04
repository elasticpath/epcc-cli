package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/google/uuid"
	"github.com/iancoleman/strcase"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/yosida95/uritemplate/v3"
	"net/url"
	"regexp"
	"strings"
)

func GetSingularTypeNames(types []string) []string {
	var ret []string

	for _, t := range types {

		otherType, ok := resources.GetResourceByName(t)

		if !ok {
			log.Warnf("Error processing resource, could not find type %s", t)
		}

		ret = append(ret, otherType.SingularName)
	}

	return ret
}

func GetParametersForTypes(types []string) string {
	r := ""

	for _, t := range types {
		r += fmt.Sprintf(" %s_ID", strings.ToUpper(t))

	}

	return r
}

func GetParameterUsageForTypes(types []string) string {
	r := ""

	for _, t := range types {
		r += fmt.Sprintf("%-20s - An ID or alias for a %s\n", t, strings.Title(t))
	}

	return r
}

func GetArgumentExampleWithIds(types []string) string {
	r := ""

	for i := 0; i < len(types); i++ {
		r += uuid.New().String() + " "
	}

	return r
}

func GetArgumentExampleWithAlias(types []string) string {
	r := ""

	for i := 0; i < len(types); i++ {
		r += "last_read=entity "
	}

	return r
}

func GetHelpResourceUrls(resourceUrl string) string {

	template, err := uritemplate.New(resourceUrl)

	if err != nil {
		return fmt.Sprintf("error: %s", err)
	}

	values := uritemplate.Values{}

	for _, varName := range template.Varnames() {
		res, ok := resources.GetResourceByName(resources.ConvertUriTemplateValueToType(varName))

		if !ok {
			values[varName] = uritemplate.String("unknown_resource:" + varName)
			continue
		}

		typeName := res.SingularName
		typeName = strings.ReplaceAll(typeName, "-", " ")
		typeName = strings.Title(typeName)
		typeName = strings.ReplaceAll(typeName, " ", "")
		typeName = strings.ReplaceAll(typeName, "V2", "")
		typeName = strcase.ToLowerCamel(typeName)

		values[varName] = uritemplate.String(":" + typeName + "Id")

	}

	templateUrl, err := template.Expand(values)

	templateUrl, _ = url.PathUnescape(templateUrl)

	return templateUrl

	//// Convert _ to " "
	//resourceUrl = strings.ReplaceAll(resourceUrl, "_", " ")
	//
	//// This will essentially snakeCase It
	//resourceUrl = strings.Title(resourceUrl)
	//
	//// Remove Spaces
	//resourceUrl = strings.ReplaceAll(resourceUrl, " ", "")
	//
	//// Use leading :
	//resourceUrl = strings.ReplaceAll(resourceUrl, "{", ":")
	//
	//// Replace closing } with Id
	//resourceUrl = strings.ReplaceAll(resourceUrl, "}", "Id")
	//
	//return resourceUrl
}

func GetArgsFunctionForResource(singularTypeNames []string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {

		var missingArgs []string

		for i, neededType := range singularTypeNames {
			if len(args) < i+1 {
				missingArgs = append(missingArgs, fmt.Sprintf("%s_ID", strings.ToUpper(neededType)))

			}
		}

		if len(missingArgs) > 0 {
			return fmt.Errorf("missing required arguments: %s must be specified, please see --help for more info", strings.Join(missingArgs, ", "))
		} else {
			return nil
		}
	}
}

var NonAlphaCharacter = regexp.MustCompile("[^A-Za-z]+")

func GetJsonKeyValuesForUsage(resource resources.Resource) string {
	var ret = ""
	for k := range resource.Attributes {

		jsonKey := k
		// A good example of why these are needed are pcm-products and the regex attributes
		jsonKey = strings.ReplaceAll(jsonKey, "^", "")
		jsonKey = strings.ReplaceAll(jsonKey, "$", "")
		jsonKey = strings.ReplaceAll(jsonKey, "\\.", ".")
		jsonKey = strings.ReplaceAll(jsonKey, "\\", "")

		jsonKey = strings.ReplaceAll(jsonKey, "([a-zA-Z0-9-_]+)", "*")
		value := strings.Trim(NonAlphaCharacter.ReplaceAllString(strings.ToUpper(k), "_"), "_ ")
		value = strings.ReplaceAll(value, "A_Z", "")
		value = strings.ReplaceAll(value, "__", "_")
		ret += " [" + jsonKey + " " + value + "]"
	}

	return ret
}

func GetJsonExample(description string, call string, header string, jsonTxt string) string {

	jsonTxt = "> " + json.PrettyPrint(jsonTxt)
	jsonTxt = strings.ReplaceAll(jsonTxt, "\n", "\n  > ")

	return fmt.Sprintf(`
  %s
   %s
  %s
  %s
`, description, call, header, jsonTxt)
}

func FillUrlWithIds(urlInfo *resources.CrudEntityInfo) string {
	var ids []string

	idsNeeded, err := resources.GetNumberOfVariablesNeeded(urlInfo.Url)

	if err != nil {
		log.Errorf("error generating help screen %v", err)
	}

	for i := 0; i < idsNeeded; i++ {
		ids = append(ids, uuid.New().String())
	}

	url, err := resources.GenerateUrl(urlInfo, ids)

	if err != nil {
		log.Errorf("error generating help screen %v", err)
	}

	return url
}
