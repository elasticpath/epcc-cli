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
	"strings"
	"sync"
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

	types, err := resources.GetTypesOfVariablesNeeded(resourceUrl)

	if err != nil {
		return fmt.Sprintf("Could not generate usage string: %s", err)
	}

	singularTypeNames := GetSingularTypeNames(types)
	parametersLongUsage := GetParameterUsageForTypes(singularTypeNames)

	return fmt.Sprintf(`Retrieves %s %s defined in a store/organization by calling %s.

%s
`, usageGetType, resourceName, GetHelpResourceUrls(resourceUrl), parametersLongUsage)
}

func GetCreateLong(resource resources.Resource) string {
	resourceName := resource.SingularName

	singularTypeNames, err := resources.GetSingularTypesOfVariablesNeeded(resource.CreateEntityInfo.Url)

	if err != nil {
		return fmt.Sprintf("Could not generate usage string: %s", err)
	}

	parametersLongUsage := GetParameterUsageForTypes(singularTypeNames)

	argumentsBlurb := ""
	switch resource.CreateEntityInfo.ContentType {
	case "multipart/form-data":
		argumentsBlurb = "Key and values are passed in using multipart/form-data encoding\n\nDocumentation:\n  " + resource.CreateEntityInfo.Docs
	case "application/json", "":
		argumentsBlurb = fmt.Sprintf(`
Key and value pairs passed in will be converted to JSON with a jq like syntax.

The EPCC CLI will automatically determine appropriate wrapping

Basic Types:
key b => { "a": "b" }
key 1 => { "a": 1  }
key '"1"' => { "a": "1" }
key true => { "a": true }
key null => { "a": null }
key '"null"'' => { "a": "null" }



Documentation:
 %s
`, resource.CreateEntityInfo.Docs)
	default:
		argumentsBlurb = fmt.Sprintf("This resource uses %s encoding, which this help doesn't know how to help you with :) Submit a bug please.\nDocumentation:\n  %s", resource.CreateEntityInfo.ContentType, resource.CreateEntityInfo.Docs)
	}

	return fmt.Sprintf(`Creates a %s in a store/organization by calling %s.
%s
%s
`, resourceName, GetHelpResourceUrls(resource.CreateEntityInfo.Url), parametersLongUsage, argumentsBlurb)
}

func GetUpdateLong(resource resources.Resource) string {
	resourceName := resource.SingularName

	singularTypeNames, err := resources.GetSingularTypesOfVariablesNeeded(resource.UpdateEntityInfo.Url)

	if err != nil {
		return fmt.Sprintf("Could not generate usage string: %s", err)
	}

	parametersLongUsage := GetParameterUsageForTypes(singularTypeNames)

	argumentsBlurb := ""
	switch resource.UpdateEntityInfo.ContentType {
	case "multipart/form-data":
		argumentsBlurb = "Key and values are passed in using multipart/form-data encoding\n\nDocumentation:\n  " + resource.DeleteEntityInfo.Docs
	case "application/json", "":
		argumentsBlurb = fmt.Sprintf(`
Key and value pairs passed in will be converted to JSON with a jq like syntax.

The EPCC CLI will automatically determine appropriate wrapping

Basic Types:
key b => { "a": "b" }
key 1 => { "a": 1  }
key '"1"' => { "a": "1" }
key true => { "a": true }
key null => { "a": null }
key '"null"'' => { "a": "null" }



Documentation:
 %s
`, resource.UpdateEntityInfo.Docs)
	default:
		argumentsBlurb = fmt.Sprintf("This resource uses %s encoding, which this help doesn't know how to help you with :) Submit a bug please.\nDocumentation:\n  %s", resource.UpdateEntityInfo.ContentType, resource.UpdateEntityInfo.Docs)
	}

	return fmt.Sprintf(`Updates a %s in a store/organization by calling %s.
%s
%s
`, resourceName, GetHelpResourceUrls(resource.UpdateEntityInfo.Url), parametersLongUsage, argumentsBlurb)
}

func GetDeleteLong(resource resources.Resource) string {
	resourceName := resource.SingularName

	singularTypeNames, err := resources.GetSingularTypesOfVariablesNeeded(resource.DeleteEntityInfo.Url)

	if err != nil {
		return fmt.Sprintf("Could not generate usage string: %s", err)
	}

	parametersLongUsage := GetParameterUsageForTypes(singularTypeNames)

	argumentsBlurb := ""
	switch resource.DeleteEntityInfo.ContentType {
	case "multipart/form-data":
		argumentsBlurb = "Key and values are passed in using multipart/form-data encoding\n\nDocumentation:\n  " + resource.DeleteEntityInfo.Docs
	case "application/json", "":
		argumentsBlurb = fmt.Sprintf(`
Key and value pairs passed in will be converted to JSON with a jq like syntax.

The EPCC CLI will automatically determine appropriate wrapping

Basic Types:
key b => { "a": "b" }
key 1 => { "a": 1  }
key '"1"' => { "a": "1" }
key true => { "a": true }
key null => { "a": null }
key '"null"'' => { "a": "null" }



Documentation:
 %s
`, resource.DeleteEntityInfo.Docs)
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

	cacheKey := fmt.Sprintf("%s-%d", resourceName, completionVerb)
	if example, ok := getExampleCache.Load(cacheKey); ok {
		return example.(string)
	}

	singularTypeNames, err := resources.GetSingularTypesOfVariablesNeeded(resourceUrl)

	if err != nil {
		return fmt.Sprintf("Could not generate example: %s", err)
	}

	exampleWithIds := fmt.Sprintf("  epcc get %s %s", resourceName, GetArgumentExampleWithIds(singularTypeNames))
	exampleWithAliases := fmt.Sprintf("  epcc get %s %s", resourceName, GetArgumentExampleWithAlias(singularTypeNames))

	examples := fmt.Sprintf("  # Retrieve %s %s\n%s\n  > GET %s\n\n", usageGetType, resourceName, exampleWithIds, FillUrlWithIds(urlInfo))

	if len(singularTypeNames) > 0 {
		examples += fmt.Sprintf("  # Retrieve %s %s using aliases \n%s\n  > GET %s\n\n", usageGetType, resourceName, exampleWithAliases, FillUrlWithIds(urlInfo))
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
			examples += fmt.Sprintf("  # Retrieve %s %s with page[limit] = 25 and page[offset] = 500 \n%s %s %s %s %s \n > GET %s \n\n", usageGetType, resourceName, exampleWithAliases, qp, "25", "page[offset]", "500", FillUrlWithIds(urlInfo)+"?page[limit]=25&page[offset]=500")

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
					examples += fmt.Sprintf("  # Retrieve %s %s sorted in ascending order of %s\n%s %s %s \n > GET %s\n\n", usageGetType, resourceName, v, exampleWithAliases, qp, v, FillUrlWithIds(urlInfo)+"?sort="+v)
				} else {
					examples += fmt.Sprintf("  # Retrieve %s %s sorted in descending order of %s\n%s %s -- %s\n > GET %s\n\n", usageGetType, resourceName, v, exampleWithAliases, qp, v, FillUrlWithIds(urlInfo)+"?sort="+v)
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
				Attributes: map[string]int{},
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

`, usageGetType, resourceName, searchOps[i], v, exampleWithAliases, qp, searchOps[i], v, FillUrlWithIds(urlInfo)+fmt.Sprintf(`?filter=%s(%s,"Hello World")`, searchOps[i], v))

				if i >= 2 {
					// Only need three examples for sort
					break
				}
			}

		default:

			examples += fmt.Sprintf("  # Retrieve %s %s with a(n) %s = %s\n%s %s %s \n > GET %s \n\n", usageGetType, resourceName, qp, "x", exampleWithAliases, qp, "x", FillUrlWithIds(urlInfo)+"?"+qp+"=x")
		}

	}

	example := strings.ReplaceAll(strings.Trim(examples, "\n"), "  ", " ")

	getExampleCache.Store(cacheKey, example)

	return example
}

var createExampleCache sync.Map

func GetCreateExample(resource resources.Resource) string {
	resourceName := resource.SingularName

	if v, ok := createExampleCache.Load(resourceName); ok {
		return v.(string)
	}

	singularTypeNames, err := resources.GetSingularTypesOfVariablesNeeded(resource.CreateEntityInfo.Url)

	if err != nil {
		return fmt.Sprintf("Could not generate example: %s", err)
	}

	exampleWithIds := fmt.Sprintf("  epcc create %s %s", resourceName, GetArgumentExampleWithIds(singularTypeNames))
	exampleWithAliases := fmt.Sprintf("  epcc create %s %s", resourceName, GetArgumentExampleWithAlias(singularTypeNames))

	baseJsonArgs := []string{}
	if !resource.NoWrapping {
		baseJsonArgs = append(baseJsonArgs, "type", resource.JsonApiType)
	}

	emptyJson, _ := json.ToJson(baseJsonArgs, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes)

	examples := GetJsonExample(fmt.Sprintf("# Create a %s", resource.SingularName), exampleWithIds, fmt.Sprintf("> POST %s", FillUrlWithIds(resource.CreateEntityInfo)), emptyJson)

	if len(singularTypeNames) > 0 {
		examples += GetJsonExample(fmt.Sprintf("# Create a %s using aliases", resource.SingularName), exampleWithIds, fmt.Sprintf("> POST %s", FillUrlWithIds(resource.CreateEntityInfo)), emptyJson)
	}

	if resource.CreateEntityInfo.ContentType != "multipart/form-data" {
		for k := range resource.Attributes {

			if k[0] == '^' {
				continue
			}

			results, _ := completion.Complete(completion.Request{
				Type:       completion.CompleteAttributeValue,
				Resource:   resource,
				Verb:       completion.Create,
				Attribute:  k,
				ToComplete: "",
			})

			arg := `"Hello World"`

			if len(results) > 0 {
				arg = results[0]
			}

			extendedArgs := append(baseJsonArgs, k, arg)

			// Don't try and use more than one key as some are mutually exclusive and the JSON will crash.
			// Resources that are heterogenous and can have array or object fields at some level (i.e., data[n].id and data.id) are examples
			jsonTxt, _ := json.ToJson(extendedArgs, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes)
			examples += GetJsonExample(fmt.Sprintf("# Create a %s passing in an argument", resourceName), fmt.Sprintf("%s %s %s", exampleWithAliases, k, arg), fmt.Sprintf("> POST %s", FillUrlWithIds(resource.CreateEntityInfo)), jsonTxt)

			autofilledData := autofill.GetJsonArrayForResource(&resource)

			extendedArgs = append(autofilledData, extendedArgs...)

			jsonTxt, _ = json.ToJson(extendedArgs, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes)
			examples += GetJsonExample(fmt.Sprintf("# Create a %s (using --auto-fill) and passing in an argument", resourceName), fmt.Sprintf("%s --auto-fill %s %s", exampleWithAliases, k, arg), fmt.Sprintf("> POST %s", FillUrlWithIds(resource.CreateEntityInfo)), jsonTxt)

			break
		}
	}

	example := strings.ReplaceAll(strings.Trim(examples, "\n"), "  ", " ")

	createExampleCache.Store(resourceName, example)

	return example
}

var updateExampleCache sync.Map

func GetUpdateExample(resource resources.Resource) string {
	resourceName := resource.SingularName

	if v, ok := updateExampleCache.Load(resourceName); ok {
		return v.(string)
	}
	singularTypeNames, err := resources.GetSingularTypesOfVariablesNeeded(resource.UpdateEntityInfo.Url)

	if err != nil {
		return fmt.Sprintf("Could not generate example: %s", err)
	}

	exampleWithIds := fmt.Sprintf("  epcc update %s %s", resourceName, GetArgumentExampleWithIds(singularTypeNames))
	exampleWithAliases := fmt.Sprintf("  epcc update %s %s", resourceName, GetArgumentExampleWithAlias(singularTypeNames))

	baseJsonArgs := []string{}
	if !resource.NoWrapping {
		baseJsonArgs = append(baseJsonArgs, "type", resource.JsonApiType)
	}

	emptyJson, _ := json.ToJson(baseJsonArgs, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes)

	examples := GetJsonExample(fmt.Sprintf("# Update a %s", resource.SingularName), exampleWithIds, fmt.Sprintf("> PUT %s", FillUrlWithIds(resource.UpdateEntityInfo)), emptyJson)

	if len(singularTypeNames) > 0 {
		examples += GetJsonExample(fmt.Sprintf("# Update a %s using aliases", resource.SingularName), exampleWithIds, fmt.Sprintf("> PUT %s", FillUrlWithIds(resource.UpdateEntityInfo)), emptyJson)
	}

	if resource.UpdateEntityInfo.ContentType != "multipart/form-data" {
		for k := range resource.Attributes {

			if k[0] == '^' {
				continue
			}

			results, _ := completion.Complete(completion.Request{
				Type:       completion.CompleteAttributeValue,
				Resource:   resource,
				Verb:       completion.Update,
				Attribute:  k,
				ToComplete: "",
			})

			arg := `"Hello World"`

			if len(results) > 0 {
				arg = results[0]
			}

			extendedArgs := append(baseJsonArgs, k, arg)

			// Don't try and use more than one key as some are mutually exclusive and the JSON will crash.
			// Resources that are heterogenous and can have array or object fields at some level (i.e., data[n].id and data.id) are examples
			jsonTxt, _ := json.ToJson(extendedArgs, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes)
			examples += GetJsonExample(fmt.Sprintf("# update a %s passing in an argument", resourceName), fmt.Sprintf("%s %s %s", exampleWithAliases, k, arg), fmt.Sprintf("> PUT %s", FillUrlWithIds(resource.UpdateEntityInfo)), jsonTxt)

			break
		}
	}

	example := strings.ReplaceAll(strings.Trim(examples, "\n"), "  ", " ")
	updateExampleCache.Store(resourceName, example)
	return example
}

var deleteExampleCache sync.Map

func GetDeleteExample(resource resources.Resource) string {
	resourceName := resource.SingularName
	if v, ok := deleteExampleCache.Load(resourceName); ok {
		return v.(string)
	}

	singularTypeNames, err := resources.GetSingularTypesOfVariablesNeeded(resource.DeleteEntityInfo.Url)

	if err != nil {
		return fmt.Sprintf("Could not generate example: %s", err)
	}

	exampleWithIds := fmt.Sprintf("  epcc delete %s %s", resourceName, GetArgumentExampleWithIds(singularTypeNames))
	exampleWithAliases := fmt.Sprintf("  epcc delete %s %s", resourceName, GetArgumentExampleWithAlias(singularTypeNames))

	baseJsonArgs := []string{}
	if !resource.NoWrapping {
		baseJsonArgs = append(baseJsonArgs, "type", resource.JsonApiType)
	}

	emptyJson, _ := json.ToJson(baseJsonArgs, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes)

	examples := GetJsonExample(fmt.Sprintf("# Delete a %s", resource.SingularName), exampleWithIds, fmt.Sprintf("> PUT %s", FillUrlWithIds(resource.DeleteEntityInfo)), emptyJson)

	if len(singularTypeNames) > 0 {
		examples += GetJsonExample(fmt.Sprintf("# Delete a %s using aliases", resource.SingularName), exampleWithIds, fmt.Sprintf("> PUT %s", FillUrlWithIds(resource.DeleteEntityInfo)), emptyJson)
	}

	if resource.DeleteEntityInfo.ContentType != "multipart/form-data" {
		for k := range resource.Attributes {

			if k[0] == '^' {
				continue
			}

			results, _ := completion.Complete(completion.Request{
				Type:       completion.CompleteAttributeValue,
				Resource:   resource,
				Verb:       completion.Delete,
				Attribute:  k,
				ToComplete: "",
			})

			arg := `"Hello World"`

			if len(results) > 0 {
				arg = results[0]
			}

			extendedArgs := append(baseJsonArgs, k, arg)

			// Don't try and use more than one key as some are mutually exclusive and the JSON will crash.
			// Resources that are heterogenous and can have array or object fields at some level (i.e., data[n].id and data.id) are examples
			jsonTxt, _ := json.ToJson(extendedArgs, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes)
			examples += GetJsonExample(fmt.Sprintf("# delete a %s passing in an argument", resourceName), fmt.Sprintf("%s %s %s", exampleWithAliases, k, arg), fmt.Sprintf("> DELETE %s", FillUrlWithIds(resource.DeleteEntityInfo)), jsonTxt)

			break
		}
	}

	example := strings.ReplaceAll(strings.Trim(examples, "\n"), "  ", " ")

	deleteExampleCache.Store(resourceName, example)

	return example
}
