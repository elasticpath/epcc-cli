package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/autofill"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/google/uuid"
	"github.com/iancoleman/strcase"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/yosida95/uritemplate/v3"
	"math/rand"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var DisableLongOutput = false
var DisableExampleOutput = false

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

func ConvertSingularTypeToCmdArg(typeName string) string {
	return fmt.Sprintf("%s_ID", strings.ReplaceAll(strings.ToUpper(typeName), "-", "_"))
}
func GetParametersForTypes(types []string) string {
	r := ""

	for _, t := range types {
		r += " " + ConvertSingularTypeToCmdArg(t)

	}

	return r
}

func GetParameterUsageForTypes(resource resources.Resource, types []string, bodyParameters bool) string {

	usage := strings.Builder{}

	if len(types) > 0 {

		usage.WriteString("Parent Resource ID Parameters (Mandatory):\n")

		for _, t := range types {
			usage.WriteString(fmt.Sprintf("  %-20s - An ID or alias for a %s\n", t, strings.Title(t)))
		}
	}

	if bodyParameters {

		usage.WriteString("\nBody Parameters (Specified as space separated key value pairs):\n")

		keys := []string{}
		maxLengthKey := 0
		for k := range resource.Attributes {
			keys = append(keys, k)

			maxLengthKey = max(maxLengthKey, len(k))
		}

		sort.Strings(keys)
		for _, k := range keys {
			v := resource.Attributes[k]

			value := "Unknown:" + v.Type

			//"pattern": "^(||RESOURCE_ID:([a-z0-9-]+|[*]))$"

			if v.Usage != "" {
				value = v.Usage
			} else if v.Type == "BOOL" {
				value = "A boolean value"
			} else if v.Type == "STRING" {
				value = "A string value"
			} else if strings.HasPrefix(v.Type, "ENUM:") {
				value = "One of the following values: " + strings.ReplaceAll(strings.ReplaceAll(v.Type, "ENUM:", ""), ",", ", ")
			} else if strings.HasPrefix(v.Type, "CONST:") {
				value = "Only: " + strings.ReplaceAll(strings.ReplaceAll(v.Type, "CONST:", ""), ",", ", ") + " (note: the epcc will auto-populate this if an adjacent attribute is set)"
			} else if v.Type == "INT" {
				value = "An integer value"
			} else if v.Type == "FLOAT" {
				value = "A floating point value"
			} else if v.Type == "URL" {
				value = "A url"
			} else if v.Type == "JSON_API_TYPE" {
				value = "A value that matches a `type` used by the API"
			} else if v.Type == "CURRENCY" {
				value = "A three letter currency code"
			} else if v.Type == "FILE" {
				value = "A filename"
			} else if v.Type == "PRIMITIVE" {
				value = "Any of an int, float, string, or boolean value"
			} else if v.Type == "SINGULAR_RESOURCE_TYPE" {
				value = "A resource name used by the epcc cli"
			} else if strings.HasPrefix(v.Type, "RESOURCE_ID") {

				resName := strings.ReplaceAll(v.Type, "RESOURCE_ID:", "")

				if res, ok := resources.GetResourceByName(resName); ok {
					attribute := "id"

					if v.AliasAttribute != "" {
						attribute = v.AliasAttribute
					}

					value = fmt.Sprintf("The %s of a %s resource", attribute, res.SingularName)
				} else {
					value = "A resource id for " + resName
				}
			}

			usage.WriteString(fmt.Sprintf("  %-"+strconv.Itoa(maxLengthKey)+"s -  %s\n", k, value))

		}

		usage.WriteString(`
Notes:
  - Other keys and values will work fine (e.g., if you are using an older version of this tool, and new features have been developed), or you have defined flows.
  - Keys with an [n] in them are array parameters and should be supplied with a [0], [1], [2], etc...
  - Mandatory body parameters are enforced by the API, and not this tool.`)
	}

	return usage.String()
}

func GetUuidsForTypes(types []string) []string {
	r := []string{}

	for i := 0; i < len(types); i++ {
		r = append(r, uuid.New().String())
	}

	return r
}

func GetArgumentExampleWithIds(types []string, uuids []string) string {
	r := ""

	for i := 0; i < len(types); i++ {
		r += uuids[i]
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
}

func GetArgFunctionForCreate(resource resources.Resource) func(cmd *cobra.Command, args []string) error {
	return GetArgFunctionForUrl(resource.SingularName, resource.CreateEntityInfo.Url)
}

func GetArgFunctionForUpdate(resource resources.Resource) func(cmd *cobra.Command, args []string) error {
	return GetArgFunctionForUrl(resource.SingularName, resource.UpdateEntityInfo.Url)
}

func GetArgFunctionForDelete(resource resources.Resource) func(cmd *cobra.Command, args []string) error {
	return GetArgFunctionForUrl(resource.SingularName, resource.DeleteEntityInfo.Url)
}

func GetArgFunctionForUrl(name, resourceUrl string) func(cmd *cobra.Command, args []string) error {

	singularTypeNames, err := resources.GetSingularTypesOfVariablesNeeded(resourceUrl)

	if err != nil {
		log.Warnf("Could not generate usage string for %s, error %v", name, err)
	}

	return func(cmd *cobra.Command, args []string) error {
		var missingArgs []string

		for i, neededType := range singularTypeNames {
			if len(args) < i+1 {
				missingArgs = append(missingArgs, ConvertSingularTypeToCmdArg(neededType))

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

	keys := []string{}
	for k := range resource.Attributes {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		v := resource.Attributes[k]

		jsonKey := k
		// A good example of why these are needed are pcm-products and the regex attributes
		jsonKey = strings.ReplaceAll(jsonKey, "^", "")
		jsonKey = strings.ReplaceAll(jsonKey, "$", "")
		jsonKey = strings.ReplaceAll(jsonKey, "\\.", ".")
		jsonKey = strings.ReplaceAll(jsonKey, "\\", "")

		jsonKey = strings.ReplaceAll(jsonKey, "([a-zA-Z0-9-_]+)", "*")
		value := "Unknown"

		// "pattern": "^(STRING|URL|INT|CONST:[a-z0-9A-Z._-]+|ENUM:[a-z0-9A-Z._,-]+|FLOAT|BOOL|FILE|CURRENCY|SINGULAR_RESOURCE_TYPE|JSON_API_TYPE|RESOURCE_ID:([a-z0-9-]+|[*]))$"

		// 	// Use is the one-line usage message.
		//	// Recommended syntax is as follows:
		//	//   [ ] identifies an optional argument. Arguments that are not enclosed in brackets are required.
		//	//   ... indicates that you can specify multiple values for the previous argument.
		//	//   |   indicates mutually exclusive information. You can use the argument to the left of the separator or the
		//	//       argument to the right of the separator. You cannot use both arguments in a single use of the command.
		//	//   { } delimits a set of mutually exclusive arguments when one of the arguments is required. If the arguments are
		//	//       optional, they are enclosed in brackets ([ ]).
		//	// Example: add [-F file | -D dir]... [-f format] profile

		if v.Type == "BOOL" {
			value = "{true|false}"
		} else if strings.HasPrefix(v.Type, "CONST:") {
			value = strings.Trim(v.Type, " ")
			value = strings.ReplaceAll(value, "CONST:", "")
		} else if strings.HasPrefix(v.Type, "ENUM:") {
			value = strings.Trim(v.Type, " ")
			value = "{" + strings.ReplaceAll(value, "ENUM:", "") + "}"
		} else {
			value = strings.Trim(NonAlphaCharacter.ReplaceAllString(strings.ToUpper(k), "_"), "_ ")
			value = strings.ReplaceAll(value, "A_Z", "")
			value = strings.ReplaceAll(value, "__", "_")
		}

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

func FillUrlWithIds(urlInfo *resources.CrudEntityInfo, uuids []string) string {
	var ids []string

	idsNeeded, err := resources.GetNumberOfVariablesNeeded(urlInfo.Url)

	if err != nil {
		log.Errorf("error generating help screen %v", err)
	}

	for i := 0; i < idsNeeded; i++ {
		ids = append(ids, uuids[i])
	}

	url, err := resources.GenerateUrl(urlInfo, ids, false)

	if err != nil {
		log.Errorf("error generating help screen %v", err)
	}

	return url
}

func GetGetShort(resourceUrl string) string {
	return fmt.Sprintf("Calls GET %s", GetHelpResourceUrls(resourceUrl))
}
func GetCreateShort(resource resources.Resource) string {
	return fmt.Sprintf("Calls POST %s", GetHelpResourceUrls(resource.CreateEntityInfo.Url))
}

func GetUpdateShort(resource resources.Resource) string {
	return fmt.Sprintf("Calls PUT %s", GetHelpResourceUrls(resource.UpdateEntityInfo.Url))
}

func GetDeleteShort(resource resources.Resource) string {
	return fmt.Sprintf("Calls DELETE %s", GetHelpResourceUrls(resource.DeleteEntityInfo.Url))
}

func GetDeleteAllShort(resource resources.Resource) string {
	return fmt.Sprintf("Calls DELETE %s for every resource in GET %s", GetHelpResourceUrls(resource.DeleteEntityInfo.Url), GetHelpResourceUrls(resource.GetCollectionInfo.Url))
}

func GetGetLong(resourceName string, resourceUrl string, usageGetType string, completionVerb int, urlInfo *resources.CrudEntityInfo, resource resources.Resource) string {

	if DisableLongOutput {
		return ""
	}

	types, err := resources.GetTypesOfVariablesNeeded(resourceUrl)

	if err != nil {
		return fmt.Sprintf("Could not generate usage string: %s", err)
	}

	singularTypeNames := GetSingularTypeNames(types)
	parametersLongUsage := GetParameterUsageForTypes(resource, singularTypeNames, false)

	return fmt.Sprintf(`Retrieves %s %s defined in a store/organization by calling %s.

%s
`, usageGetType, resourceName, GetHelpResourceUrls(resourceUrl), parametersLongUsage)
}

func GetJsonSyntaxExample(resource resources.Resource, verb string, parentIds string, id string) string {
	allIds := parentIds + id

	prefix := fmt.Sprintf(`
Key Value pairs passed in will be converted to JSON with a jq like syntax.

The EPCC CLI will automatically determine appropriate wrapping (i.e., wrap the values in a data key or attributes key)

The following examples show JSON syntax (and also include required parent ids)
# Simple type with key and value 
epcc %s %s%skey value => %s

# Numeric types will be encoded as json numbers
epcc %s %s%skey 1 => %s

# If a value *must* be a string, you should wrap it in quotes, be mindful that your shell may require you to quote quotes :)
epcc %s %s%skey '"1"' => %s

# If a value starts with a -, you should place a -- somewhere in the string before hand, this will turn off flag intepretation
epcc %s %s%skey -- -value => %s 

# Boolean types work similarly
epcc %s %s%skey true => %s

# As does null
epcc %s %s%skey null => %s

# Which can be encoded with quotes
epcc %s %s%skey '"null"' => %s

# To send an array use the following syntax
epcc %s %s%skey[0] a key[1] true => %s

# To send an empty array use the following syntax (apologies)
epcc %s %s%skey [] => %s

# To send a nested object use the . character to nest values deeper.
epcc %s %s%skey.some.child hello key.some.other goodbye => %s

# Attributes can also be generated using Go templates and Sprig (https://masterminds.github.io/sprig/) functions.
epcc %s %s%skey 'Test {{ randAlphaNum 6 | upper }} Value' => %s`,
		verb, resource.SingularName, allIds, toJsonExample([]string{"key", "b"}, resource),
		verb, resource.SingularName, allIds, toJsonExample([]string{"key", "1"}, resource),
		verb, resource.SingularName, allIds, toJsonExample([]string{"key", "\"1\""}, resource),
		verb, resource.SingularName, allIds, toJsonExample([]string{"key", "-value"}, resource),
		verb, resource.SingularName, allIds, toJsonExample([]string{"key", "true"}, resource),
		verb, resource.SingularName, allIds, toJsonExample([]string{"key", "null"}, resource),
		verb, resource.SingularName, allIds, toJsonExample([]string{"key", "\"null\""}, resource),
		verb, resource.SingularName, allIds, toJsonExample([]string{"key[0]", "a", "key[1]", "true"}, resource),
		verb, resource.SingularName, allIds, toJsonExample([]string{"key", "[]"}, resource),
		verb, resource.SingularName, allIds, toJsonExample([]string{"key.some.child", "hello", "key.some.other", "goodbye"}, resource),
		verb, resource.SingularName, allIds, toJsonExample([]string{"key", "Test {{ randAlphaNum 6 | upper }} Value"}, resource),
	)

	return prefix
}

func toJsonExample(in []string, resource resources.Resource) string {

	if !resource.NoWrapping {
		in = append([]string{"type", resource.JsonApiType}, in...)
	}

	jsonTxt, err := json.ToJson(in, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes, false, true)

	if err != nil {
		return fmt.Sprintf("Could not get json: %s", err)
	}

	return jsonTxt
}

func GetCreateLong(resource resources.Resource) string {
	if DisableLongOutput {
		return ""
	}
	resourceName := resource.SingularName

	singularTypeNames, err := resources.GetSingularTypesOfVariablesNeeded(resource.CreateEntityInfo.Url)

	if err != nil {
		return fmt.Sprintf("Could not generate usage string: %s", err)
	}

	parametersLongUsage := GetParameterUsageForTypes(resource, singularTypeNames, true)

	uuids := GetUuidsForTypes(singularTypeNames)
	exampleWithIds := " " + GetArgumentExampleWithIds(singularTypeNames, uuids)

	argumentsBlurb := ""
	switch resource.CreateEntityInfo.ContentType {
	case "multipart/form-data":
		argumentsBlurb = "Key and values are passed in using multipart/form-data encoding\n\nDocumentation:\n  " + resource.CreateEntityInfo.Docs
	case "application/json", "":
		argumentsBlurb = fmt.Sprintf(`
%s

Documentation:
 %s
`, GetJsonSyntaxExample(resource, "create", exampleWithIds, ""), resource.CreateEntityInfo.Docs)
	default:
		argumentsBlurb = fmt.Sprintf("This resource uses %s encoding, which this help doesn't know how to help you with :) Submit a bug please.\nDocumentation:\n  %s", resource.CreateEntityInfo.ContentType, resource.CreateEntityInfo.Docs)
	}

	return fmt.Sprintf(`Creates a %s in a store/organization by calling %s.
%s
%s
`, resourceName, GetHelpResourceUrls(resource.CreateEntityInfo.Url), parametersLongUsage, argumentsBlurb)
}

func GetUpdateLong(resource resources.Resource) string {
	if DisableLongOutput {
		return ""
	}
	resourceName := resource.SingularName

	singularTypeNames, err := resources.GetSingularTypesOfVariablesNeeded(resource.UpdateEntityInfo.Url)

	if err != nil {
		return fmt.Sprintf("Could not generate usage string: %s", err)
	}

	uuids := GetUuidsForTypes(singularTypeNames)
	exampleWithIds := " " + GetArgumentExampleWithIds(singularTypeNames, uuids)

	parametersLongUsage := GetParameterUsageForTypes(resource, singularTypeNames, true)

	argumentsBlurb := ""
	switch resource.UpdateEntityInfo.ContentType {
	case "multipart/form-data":
		argumentsBlurb = "Key and values are passed in using multipart/form-data encoding\n\nDocumentation:\n  " + resource.DeleteEntityInfo.Docs
	case "application/json", "":
		argumentsBlurb = fmt.Sprintf(`
%s

Documentation:
 %s
`, GetJsonSyntaxExample(resource, "update", exampleWithIds, " 00000000-feed-dada-iced-c0ffee000000"), resource.UpdateEntityInfo.Docs)
	default:
		argumentsBlurb = fmt.Sprintf("This resource uses %s encoding, which this help doesn't know how to help you with :) Submit a bug please.\nDocumentation:\n  %s", resource.UpdateEntityInfo.ContentType, resource.UpdateEntityInfo.Docs)
	}

	return fmt.Sprintf(`Updates a %s in a store/organization by calling %s.
%s
%s
`, resourceName, GetHelpResourceUrls(resource.UpdateEntityInfo.Url), parametersLongUsage, argumentsBlurb)
}

func GetDeleteLong(resource resources.Resource) string {
	if DisableLongOutput {
		return ""
	}
	resourceName := resource.SingularName

	singularTypeNames, err := resources.GetSingularTypesOfVariablesNeeded(resource.DeleteEntityInfo.Url)

	if err != nil {
		return fmt.Sprintf("Could not generate usage string: %s", err)
	}

	uuids := GetUuidsForTypes(singularTypeNames)
	exampleWithIds := " " + GetArgumentExampleWithIds(singularTypeNames, uuids)

	parametersLongUsage := GetParameterUsageForTypes(resource, singularTypeNames, false)

	argumentsBlurb := ""
	switch resource.DeleteEntityInfo.ContentType {
	case "multipart/form-data":
		argumentsBlurb = "Key and values are passed in using multipart/form-data encoding\n\nDocumentation:\n  " + resource.DeleteEntityInfo.Docs
	case "application/json", "":
		argumentsBlurb = fmt.Sprintf(`
%s

Documentation:
 %s
`, GetJsonSyntaxExample(resource, "delete", exampleWithIds, " 00000000-feed-dada-iced-c0ffee000000"), resource.DeleteEntityInfo.Docs)
	default:
		argumentsBlurb = fmt.Sprintf("This resource uses %s encoding, which this help doesn't know how to help you with :) Submit a bug please.\nDocumentation:\n  %s", resource.DeleteEntityInfo.ContentType, resource.DeleteEntityInfo.Docs)
	}

	return fmt.Sprintf(`Deletes a %s in a store/organization by calling %s.
%s
%s
`, resourceName, GetHelpResourceUrls(resource.DeleteEntityInfo.Url), parametersLongUsage, argumentsBlurb)
}

func GetGetUsageString(resourceName string, resourceUrl string, completionVerb int, resource resources.Resource) string {
	singularTypeNames, err := resources.GetSingularTypesOfVariablesNeeded(resourceUrl)

	if err != nil {
		log.Warnf("Could not generate usage string for %s, error %v", resourceName, err)
		return resourceName
	}

	usageString := resourceName + GetParametersForTypes(singularTypeNames)

	queryParameters, _ := completion.Complete(completion.Request{
		Type:     completion.CompleteQueryParamKey,
		Resource: resource,
		Verb:     completionVerb,
	})

	for _, qp := range queryParameters {
		if qp == "" {
			continue
		}

		switch qp {
		case "page[limit]":
			usageString += fmt.Sprintf(" [page[limit] N]")
		case "page[offset]":
			// No example
			usageString += fmt.Sprintf(" [page[offset] N]")
		case "page[total_method]":
			usageString += fmt.Sprintf(" [page[total_method] {exact,observed,estimate,lower_bound,cached,none}]")
		case "sort":
			usageString += fmt.Sprintf(" [sort SORT]")
		case "filter":
			usageString += fmt.Sprintf(" [filter FILTER]")
		default:
			usageString += fmt.Sprintf(" [%s VALUE]", qp)
		}

	}

	return usageString
}
func GetCreateUsageString(resource resources.Resource) string {
	resourceName := resource.SingularName

	singularTypeNames, err := resources.GetSingularTypesOfVariablesNeeded(resource.CreateEntityInfo.Url)

	if err != nil {
		log.Warnf("Could not generate usage string for %s, error %v", resourceName, err)
		return resourceName
	}

	return resourceName + GetParametersForTypes(singularTypeNames) + GetJsonKeyValuesForUsage(resource)
}

func GetUpdateUsage(resource resources.Resource) string {
	resourceName := resource.SingularName

	singularTypeNames, err := resources.GetSingularTypesOfVariablesNeeded(resource.UpdateEntityInfo.Url)

	if err != nil {
		log.Warnf("Could not generate usage string for %s, error %v", resourceName, err)
		return resourceName
	}

	return resourceName + GetParametersForTypes(singularTypeNames) + GetJsonKeyValuesForUsage(resource)
}

func GetDeleteUsage(resource resources.Resource) string {
	resourceName := resource.SingularName

	singularTypeNames, err := resources.GetSingularTypesOfVariablesNeeded(resource.DeleteEntityInfo.Url)

	if err != nil {
		log.Warnf("Could not generate usage string for %s, error %v", resourceName, err)
		return resourceName
	}

	return resourceName + GetParametersForTypes(singularTypeNames) + GetJsonKeyValuesForUsage(resource)
}

var getExampleCache sync.Map

func GetGetExample(resourceName string, resourceUrl string, usageGetType string, completionVerb int, urlInfo *resources.CrudEntityInfo, resource resources.Resource) string {

	if DisableExampleOutput {
		return ""
	}

	cacheKey := fmt.Sprintf("%s-%d", resourceName, completionVerb)
	if example, ok := getExampleCache.Load(cacheKey); ok {
		return example.(string)
	}

	singularTypeNames, err := resources.GetSingularTypesOfVariablesNeeded(resourceUrl)

	if err != nil {
		return fmt.Sprintf("Could not generate example: %s", err)
	}

	uuids := GetUuidsForTypes(singularTypeNames)
	exampleWithIds := fmt.Sprintf("  epcc get %s %s", resourceName, GetArgumentExampleWithIds(singularTypeNames, uuids))
	exampleWithAliases := fmt.Sprintf("  epcc get %s %s", resourceName, GetArgumentExampleWithAlias(singularTypeNames))

	examples := fmt.Sprintf("  # Retrieve %s %s\n%s\n  > GET %s\n\n", usageGetType, resourceName, exampleWithIds, FillUrlWithIds(urlInfo, uuids))

	if len(singularTypeNames) > 0 {
		examples += fmt.Sprintf("  # Retrieve %s %s using aliases \n%s\n  > GET %s\n\n", usageGetType, resourceName, exampleWithAliases, FillUrlWithIds(urlInfo, uuids))
	}

	queryParameters, _ := completion.Complete(completion.Request{
		Type:     completion.CompleteQueryParamKey,
		Resource: resource,
		Verb:     completionVerb,
	})

	for _, qp := range queryParameters {
		if qp == "" {
			continue
		}

		switch qp {
		case "page[limit]":
			examples += fmt.Sprintf("  # Retrieve %s %s with page[limit] = 25 and page[offset] = 500 \n%s %s %s %s %s \n > GET %s \n\n", usageGetType, resourceName, exampleWithAliases, qp, "25", "page[offset]", "500", FillUrlWithIds(urlInfo, uuids)+"?page[limit]=25&page[offset]=500")

		case "sort":

			sortKeys, _ := completion.Complete(completion.Request{
				Type:       completion.CompleteQueryParamValue,
				Resource:   resource,
				QueryParam: "sort",
				Verb:       completionVerb,
			})

			rand.Shuffle(len(sortKeys), func(i, j int) {
				sortKeys[i], sortKeys[j] = sortKeys[j], sortKeys[i]
			})

			for i, v := range sortKeys {
				if v[0] != '-' {
					examples += fmt.Sprintf("  # Retrieve %s %s sorted in ascending order of %s\n%s %s %s \n > GET %s\n\n", usageGetType, resourceName, v, exampleWithAliases, qp, v, FillUrlWithIds(urlInfo, uuids)+"?sort="+v)
				} else {
					examples += fmt.Sprintf("  # Retrieve %s %s sorted in descending order of %s\n%s %s -- %s\n > GET %s\n\n", usageGetType, resourceName, v, exampleWithAliases, qp, v, FillUrlWithIds(urlInfo, uuids)+"?sort="+v)
				}

				if i > 2 {
					// Only need three examples for sort
					break
				}
			}

		case "filter":

			attributeKeys, _ := completion.Complete(completion.Request{
				Type:       completion.CompleteAttributeKey,
				Resource:   resource,
				Attributes: map[string]struct{}{},
				Verb:       completion.Create,
			})

			rand.Shuffle(len(attributeKeys), func(i, j int) {
				attributeKeys[i], attributeKeys[j] = attributeKeys[j], attributeKeys[i]
			})

			searchOps := []string{"eq", "like", "gt"}
			for i, v := range attributeKeys {
				examples += fmt.Sprintf(`  # Retrieve %s %s with filter %s(%s,"Hello World")
  %s %s '%s(%s,"Hello World")'
 > GET %s

`, usageGetType, resourceName, searchOps[i], v, exampleWithAliases, qp, searchOps[i], v, FillUrlWithIds(urlInfo, uuids)+fmt.Sprintf(`?filter=%s(%s,"Hello World")`, searchOps[i], v))

				if i >= 2 {
					// Only need three examples for sort
					break
				}
			}

		default:

			examples += fmt.Sprintf("  # Retrieve %s %s with a(n) %s = %s\n%s %s %s \n > GET %s \n\n", usageGetType, resourceName, qp, "x", exampleWithAliases, qp, "x", FillUrlWithIds(urlInfo, uuids)+"?"+qp+"=x")
		}

	}

	example := strings.ReplaceAll(strings.Trim(examples, "\n"), "  ", " ")

	getExampleCache.Store(cacheKey, example)

	return example
}

var createExampleCache sync.Map

func GetCreateExample(resource resources.Resource) string {
	if DisableExampleOutput {
		return ""
	}

	resourceName := resource.SingularName

	if v, ok := createExampleCache.Load(resourceName); ok {
		return v.(string)
	}

	singularTypeNames, err := resources.GetSingularTypesOfVariablesNeeded(resource.CreateEntityInfo.Url)

	if err != nil {
		return fmt.Sprintf("Could not generate example: %s", err)
	}

	uuids := GetUuidsForTypes(singularTypeNames)
	exampleWithIds := fmt.Sprintf("  epcc create %s %s", resourceName, GetArgumentExampleWithIds(singularTypeNames, uuids))

	exampleWithAliases := fmt.Sprintf("  epcc create %s %s", resourceName, GetArgumentExampleWithAlias(singularTypeNames))

	baseJsonArgs := []string{}
	if !resource.NoWrapping {
		baseJsonArgs = append(baseJsonArgs, "type", resource.JsonApiType)
	}

	emptyJson, _ := json.ToJson(baseJsonArgs, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes, false, true)

	examples := GetJsonExample(fmt.Sprintf("# Create a %s", resource.SingularName), exampleWithIds, fmt.Sprintf("> POST %s", FillUrlWithIds(resource.CreateEntityInfo, uuids)), emptyJson)

	if len(singularTypeNames) > 0 {
		examples += GetJsonExample(fmt.Sprintf("# Create a %s using aliases", resource.SingularName), exampleWithIds, fmt.Sprintf("> POST %s", FillUrlWithIds(resource.CreateEntityInfo, uuids)), emptyJson)
	}

	if resource.CreateEntityInfo.ContentType != "multipart/form-data" {
		for kOrig, v := range resource.Attributes {

			if kOrig[0] == '^' {
				continue
			}

			if strings.HasPrefix(v.Type, "CONST:") {
				continue
			}

			results, _ := completion.Complete(completion.Request{
				Type:       completion.CompleteAttributeValue,
				Resource:   resource,
				Verb:       completion.Create,
				Attribute:  kOrig,
				ToComplete: "",
				NoAliases:  true,
			})

			k := strings.ReplaceAll(kOrig, "[n]", "[0]")

			arg := `"Hello World"`

			if len(results) > 0 {
				arg = results[0]
			}

			extendedArgs := append(baseJsonArgs, k, arg)

			// Don't try and use more than one key as some are mutually exclusive and the JSON will crash.
			// Resources that are heterogenous and can have array or object fields at some level (i.e., data[n].id and data.id) are examples
			jsonTxt, _ := json.ToJson(extendedArgs, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes, false, true)
			examples += GetJsonExample(fmt.Sprintf("# Create a %s passing in an argument", resourceName), fmt.Sprintf("%s %s %s", exampleWithAliases, k, arg), fmt.Sprintf("> POST %s", FillUrlWithIds(resource.CreateEntityInfo, uuids)), jsonTxt)

			autofilledData := autofill.GetJsonArrayForResource(&resource)

			extendedArgs = append(autofilledData, extendedArgs...)

			jsonTxt, _ = json.ToJson(extendedArgs, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes, false, true)
			examples += GetJsonExample(fmt.Sprintf("# Create a %s (using --auto-fill) and passing in an argument", resourceName), fmt.Sprintf("%s --auto-fill %s %s", exampleWithAliases, k, arg), fmt.Sprintf("> POST %s", FillUrlWithIds(resource.CreateEntityInfo, uuids)), jsonTxt)

			break
		}
	}

	example := strings.ReplaceAll(strings.Trim(examples, "\n"), "  ", " ")

	createExampleCache.Store(resourceName, example)

	return example
}

var updateExampleCache sync.Map

func GetUpdateExample(resource resources.Resource) string {
	if DisableExampleOutput {
		return ""
	}
	resourceName := resource.SingularName

	if v, ok := updateExampleCache.Load(resourceName); ok {
		return v.(string)
	}
	singularTypeNames, err := resources.GetSingularTypesOfVariablesNeeded(resource.UpdateEntityInfo.Url)

	if err != nil {
		return fmt.Sprintf("Could not generate example: %s", err)
	}

	uuids := GetUuidsForTypes(singularTypeNames)
	exampleWithIds := fmt.Sprintf("  epcc update %s %s", resourceName, GetArgumentExampleWithIds(singularTypeNames, uuids))
	exampleWithAliases := fmt.Sprintf("  epcc update %s %s", resourceName, GetArgumentExampleWithAlias(singularTypeNames))

	baseJsonArgs := []string{}
	if !resource.NoWrapping {
		baseJsonArgs = append(baseJsonArgs, "type", resource.JsonApiType)
	}

	emptyJson, _ := json.ToJson(baseJsonArgs, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes, false, true)

	examples := GetJsonExample(fmt.Sprintf("# Update a %s", resource.SingularName), exampleWithIds, fmt.Sprintf("> PUT %s", FillUrlWithIds(resource.UpdateEntityInfo, uuids)), emptyJson)

	if len(singularTypeNames) > 0 {
		examples += GetJsonExample(fmt.Sprintf("# Update a %s using aliases", resource.SingularName), exampleWithIds, fmt.Sprintf("> PUT %s", FillUrlWithIds(resource.UpdateEntityInfo, uuids)), emptyJson)
	}

	if resource.UpdateEntityInfo.ContentType != "multipart/form-data" {
		for kOrig, v := range resource.Attributes {

			if kOrig[0] == '^' {
				continue
			}

			if strings.HasPrefix(v.Type, "CONST:") {
				continue
			}

			results, _ := completion.Complete(completion.Request{
				Type:       completion.CompleteAttributeValue,
				Resource:   resource,
				Verb:       completion.Update,
				Attribute:  kOrig,
				ToComplete: "",
				NoAliases:  true,
			})

			k := strings.ReplaceAll(kOrig, "[n]", "[0]")

			arg := `"Hello World"`

			if len(results) > 0 {
				arg = results[0]
			}

			extendedArgs := append(baseJsonArgs, k, arg)

			// Don't try and use more than one key as some are mutually exclusive and the JSON will crash.
			// Resources that are heterogenous and can have array or object fields at some level (i.e., data[n].id and data.id) are examples
			jsonTxt, _ := json.ToJson(extendedArgs, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes, false, true)
			examples += GetJsonExample(fmt.Sprintf("# update a %s passing in an argument", resourceName), fmt.Sprintf("%s %s %s", exampleWithAliases, k, arg), fmt.Sprintf("> PUT %s", FillUrlWithIds(resource.UpdateEntityInfo, uuids)), jsonTxt)

			break
		}
	}

	example := strings.ReplaceAll(strings.Trim(examples, "\n"), "  ", " ")
	updateExampleCache.Store(resourceName, example)
	return example
}

var deleteExampleCache sync.Map

func GetDeleteExample(resource resources.Resource) string {
	if DisableExampleOutput {
		return ""
	}

	resourceName := resource.SingularName
	if v, ok := deleteExampleCache.Load(resourceName); ok {
		return v.(string)
	}

	singularTypeNames, err := resources.GetSingularTypesOfVariablesNeeded(resource.DeleteEntityInfo.Url)

	if err != nil {
		return fmt.Sprintf("Could not generate example: %s", err)
	}

	uuids := GetUuidsForTypes(singularTypeNames)
	exampleWithIds := fmt.Sprintf("  epcc delete %s %s", resourceName, GetArgumentExampleWithIds(singularTypeNames, uuids))
	exampleWithAliases := fmt.Sprintf("  epcc delete %s %s", resourceName, GetArgumentExampleWithAlias(singularTypeNames))

	baseJsonArgs := []string{}
	if !resource.NoWrapping {
		baseJsonArgs = append(baseJsonArgs, "type", resource.JsonApiType)
	}

	emptyJson, _ := json.ToJson(baseJsonArgs, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes, false, true)

	examples := GetJsonExample(fmt.Sprintf("# Delete a %s", resource.SingularName), exampleWithIds, fmt.Sprintf("> PUT %s", FillUrlWithIds(resource.DeleteEntityInfo, uuids)), emptyJson)

	if len(singularTypeNames) > 0 {
		examples += GetJsonExample(fmt.Sprintf("# Delete a %s using aliases", resource.SingularName), exampleWithIds, fmt.Sprintf("> PUT %s", FillUrlWithIds(resource.DeleteEntityInfo, uuids)), emptyJson)
	}

	if resource.DeleteEntityInfo.ContentType != "multipart/form-data" {
		for kOrig, v := range resource.Attributes {

			if kOrig[0] == '^' {
				continue
			}

			if strings.HasPrefix(v.Type, "CONST:") {
				continue
			}

			results, _ := completion.Complete(completion.Request{
				Type:       completion.CompleteAttributeValue,
				Resource:   resource,
				Verb:       completion.Delete,
				Attribute:  kOrig,
				ToComplete: "",
				NoAliases:  true,
			})

			k := strings.ReplaceAll(kOrig, "[n]", "[0]")

			arg := `"Hello World"`

			if len(results) > 0 {
				arg = results[0]
			}

			extendedArgs := append(baseJsonArgs, k, arg)

			// Don't try and use more than one key as some are mutually exclusive and the JSON will crash.
			// Resources that are heterogenous and can have array or object fields at some level (i.e., data[n].id and data.id) are examples
			jsonTxt, _ := json.ToJson(extendedArgs, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes, false, true)
			examples += GetJsonExample(fmt.Sprintf("# delete a %s passing in an argument", resourceName), fmt.Sprintf("%s %s %s", exampleWithAliases, k, arg), fmt.Sprintf("> DELETE %s", FillUrlWithIds(resource.DeleteEntityInfo, uuids)), jsonTxt)

			break
		}
	}

	example := strings.ReplaceAll(strings.Trim(examples, "\n"), "  ", " ")

	deleteExampleCache.Store(resourceName, example)

	return example
}

func getCommandForResource(cmd *cobra.Command, res string) *cobra.Command {
	for _, c := range cmd.Commands() {
		if strings.HasPrefix(c.Use, res+" ") {
			return c
		}
	}
	return nil
}
