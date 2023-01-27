package resources__test

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/santhosh-tekuri/jsonschema/v4"
	log "github.com/sirupsen/logrus"
	"github.com/yosida95/uritemplate/v3"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestUriTemplatesAllReferenceValidResource(t *testing.T) {
	// Fixture Setup

	// nothing needed.

	// Execute SUT
	errors := ""
	for key, val := range resources.GetPluralResources() {

		if val.CreateEntityInfo != nil {
			err := validateCrudEntityInfo(*val.CreateEntityInfo)
			if err != "" {
				errors += fmt.Sprintf("Could not process CREATE uri for resource `%s`, error:\n%s\n", key, err)
			}
		}

		if val.UpdateEntityInfo != nil {
			err := validateCrudEntityInfo(*val.UpdateEntityInfo)
			if err != "" {
				errors += fmt.Sprintf("Could not process UPDATE uri for resource `%s`, error:\n%s\n", key, err)
			}
		}

		if val.DeleteEntityInfo != nil {

			err := validateCrudEntityInfo(*val.DeleteEntityInfo)
			if err != "" {
				errors += fmt.Sprintf("Could not process DELETE uri for resource `%s`, error:\n%s\n", key, err)
			}
		}

		if val.GetEntityInfo != nil {
			err := validateCrudEntityInfo(*val.GetEntityInfo)
			if err != "" {
				errors += fmt.Sprintf("Could not process GET entity uri for resource `%s`, error:\n%s\n", key, err)
			}
		}

		if val.GetCollectionInfo != nil {
			err := validateCrudEntityInfo(*val.GetCollectionInfo)
			if err != "" {
				errors += fmt.Sprintf("Could not process GET collection uri for resource `%s`, error:\n%s\n", key, err)
			}
		}

		for attributeKey, attributeInfo := range val.Attributes {
			err := validateAttributeInfo(attributeInfo)
			if err != "" {
				errors += fmt.Sprintf("Couldn't process attributes for resource `%s` attribute `%s`, error:\n%s\n", key, attributeKey, err)
			}
		}
	}

	// Verification
	if len(errors) > 0 {
		t.Fatalf("Errors occurred while validating URI Templates:\n%s", errors)
	}
}

var arrayLiteralIndex = regexp.MustCompile("\\[[0-9]+]")

func validateAttributeInfo(info *resources.CrudEntityAttribute) string {
	match := arrayLiteralIndex.Match([]byte(info.Key))
	errors := ""

	if info.Key[0] == '^' {
		if info.Key[len(info.Key)-1] != '$' {
			errors += fmt.Sprintf("\t attribute `%s` starts with a ^ but doesn't end with a $, this is likely a bug due to regex rules)\n", info.Key)
		} else {
			if _, err := regexp.Compile(info.Key); err != nil {
				errors += fmt.Sprintf("\t attribute `%s` is a regex, but it doesn't compile: %v", info.Key, err)
			}

			rt := completion.NewRegexCompletionTree()
			if err := rt.AddRegex(info.Key); err != nil {
				errors += fmt.Sprintf("\t attribute `%s` is a regex, but the completion tree doesn't support it: %v", info.Key, err)
			}

		}
	}
	if match {
		errors += fmt.Sprintf("\t attribute `%s` has array index (e.g., [4] instead of [n], this is almost certainly a bug)\n", info.Key)
	}

	if strings.HasPrefix(info.Type, "RESOURCE_ID:") {
		resourceType := info.Type[len("RESOURCE_ID:"):]
		if _, ok := resources.GetResourceByName(resourceType); !ok {

			if _, ok := resources.GetSingularResourceByName(resourceType); !ok {
				errors += fmt.Sprintf("\t attribute `%s` references a resource type that doesn't exist: %s\n", info.Key, resourceType)
			}
		}
	}

	return errors
}
func validateCrudEntityInfo(info resources.CrudEntityInfo) string {
	errors := ""

	template, err := uritemplate.New(info.Url)
	if err != nil {
		errors += fmt.Sprintf("\tCould not process Uri %s for templates error:%s\n", info.Url, err)
	} else {
		variables := map[string]bool{}
		for _, variable := range template.Varnames() {
			variables[variable] = true
			resourceName := strings.ReplaceAll(variable, "_", "-")
			if _, ok := resources.GetPluralResources()[resourceName]; !ok {
				errors += fmt.Sprintf("\tError processing Uri %s, the URI template references a resource %s, but could not find it\n", info.Url, resourceName)
			}
		}

		for key, value := range info.ParentResourceValueOverrides {
			if value != "slug" && value != "sku" && value != "id" {
				errors += fmt.Sprintf("\tUrl %s has an invalid override for %s => %s\n", info.Url, key, value)
			}

			if _, ok := variables[key]; !ok {
				errors += fmt.Sprintf("\tUrl %s has an invalid override for %s, this key doesn't exist in the URL", info.Url, key)
			}
		}

	}

	return errors
}

func TestJsonSchemaValidate(t *testing.T) {
	sch, err := jsonschema.Compile("../resources_schema.json")
	if err != nil {
		log.Fatalf("%#v", err)
	}

	data, err := os.ReadFile("resources.yaml")
	if err != nil {
		log.Fatal(err)
	}

	var v interface{}
	if err := yaml.Unmarshal(data, &v); err != nil {
		log.Fatal(err)
	}

	if err = sch.ValidateInterface(v); err != nil {
		log.Fatalf("%#v", err)
	}
}

func TestResourceDocsExist(t *testing.T) {
	const httpStatusCodeOk = 200

	Resources := resources.GetPluralResources()
	linksReferenceCount := make(map[string]int, len(Resources))

	for resource := range Resources {
		linksReferenceCount[Resources[resource].Docs]++
		if Resources[resource].GetCollectionInfo != nil {
			linksReferenceCount[Resources[resource].GetCollectionInfo.Docs]++
		}
		if Resources[resource].CreateEntityInfo != nil {
			linksReferenceCount[Resources[resource].CreateEntityInfo.Docs]++
		}
		if Resources[resource].GetEntityInfo != nil {
			linksReferenceCount[Resources[resource].GetEntityInfo.Docs]++
		}
		if Resources[resource].UpdateEntityInfo != nil {
			linksReferenceCount[Resources[resource].UpdateEntityInfo.Docs]++
		}
		if Resources[resource].DeleteEntityInfo != nil {
			linksReferenceCount[Resources[resource].DeleteEntityInfo.Docs]++
		}
	}

	for link := range linksReferenceCount {
		response, err := http.DefaultClient.Head(link)
		if err != nil {
			t.Errorf("Error Retrieving Link\nLink: %s\nError Message: %s\nReference Count: %d", link, err, linksReferenceCount[link])
		} else {
			if response.StatusCode != httpStatusCodeOk {
				t.Errorf("Unexpected Response\nLink: %s\nExpected Status Code: %d\nActual Status Code: %d\nReference Count: %d",
					link, httpStatusCodeOk, response.StatusCode, linksReferenceCount[link])
			}
			if err := response.Body.Close(); err != nil {
				t.Errorf("Error Closing Reponse Body\nError Message: %s", err)
			}
		}
	}
}
