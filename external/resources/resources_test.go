package resources

import (
	"fmt"
	"github.com/santhosh-tekuri/jsonschema/v4"
	log "github.com/sirupsen/logrus"
	"github.com/yosida95/uritemplate/v3"
	"gopkg.in/yaml.v3"
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
	for key, val := range resources {

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

func validateAttributeInfo(info *CrudEntityAttribute) string {
	match := arrayLiteralIndex.Match([]byte(info.Key))
	errors := ""
	if match {
		errors += fmt.Sprintf("\t attribute `%s` has array index (e.g., [4] instead of [n], this is almost certainly a bug)\n", info.Key)
	}

	if strings.HasPrefix(info.Type, "RESOURCE_ID:") {
		resourceType := info.Type[len("RESOURCE_ID:"):]
		if _, ok := resources[resourceType]; !ok {
			if _, ok := resourcesSingular[resourceType]; !ok {
				errors += fmt.Sprintf("\t attribute `%s` references a resource type that doesn't exist: %s\n", info.Key, resourceType)
			}
		}
	}

	return errors
}
func validateCrudEntityInfo(info CrudEntityInfo) string {
	errors := ""

	template, err := uritemplate.New(info.Url)
	if err != nil {
		errors += fmt.Sprintf("\tCould not process Uri %s for templates error:%s\n", info.Url, err)
	} else {
		variables := map[string]bool{}
		for _, variable := range template.Varnames() {
			variables[variable] = true
			resourceName := strings.ReplaceAll(variable, "_", "-")
			if _, ok := resources[resourceName]; !ok {
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
	sch, err := jsonschema.Compile("resources_schema.json")
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
