package resources

import (
	"fmt"
	"github.com/yosida95/uritemplate/v3"
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
			template, err := uritemplate.New(val.CreateEntityInfo.Url)

			if err != nil {
				errors += fmt.Sprintf("Could not process CREATE uri for resource %s, error:%s\n", key, err)
			}

			for _, variable := range template.Varnames() {
				resourceName := strings.ReplaceAll(variable, "_", "-")
				if _, ok := resources[resourceName]; !ok {
					errors += fmt.Sprintf("Error processing CREATE uri for resource %s, the URI template references a resource %s, but could not find it", key, resourceName)
				}
			}
		}

		if val.UpdateEntityInfo != nil {
			template, err := uritemplate.New(val.UpdateEntityInfo.Url)

			if err != nil {
				errors += fmt.Sprintf("Could not process UPDATE uri for resource %s, error:%s\n", key, err)
			}

			for _, variable := range template.Varnames() {
				resourceName := strings.ReplaceAll(variable, "_", "-")
				if _, ok := resources[resourceName]; !ok {
					errors += fmt.Sprintf("Error processing UPDATE uri for resource %s, the URI template references a resource %s, but could not find it", key, resourceName)
				}
			}
		}

		if val.DeleteEntityInfo != nil {
			template, err := uritemplate.New(val.DeleteEntityInfo.Url)

			if err != nil {
				errors += fmt.Sprintf("Could not process DELETE uri for resource %s, error:%s\n", key, err)
			}

			for _, variable := range template.Varnames() {
				resourceName := strings.ReplaceAll(variable, "_", "-")
				if _, ok := resources[resourceName]; !ok {
					errors += fmt.Sprintf("Error processing DELETE uri for resource %s, the URI template references a resource %s, but could not find it", key, resourceName)
				}
			}
		}

		if val.GetEntityInfo != nil {
			template, err := uritemplate.New(val.GetEntityInfo.Url)

			if err != nil {
				errors += fmt.Sprintf("Could not process GET entity uri for resource %s, error:%s\n", key, err)
			}

			for _, variable := range template.Varnames() {
				resourceName := strings.ReplaceAll(variable, "_", "-")
				if _, ok := resources[resourceName]; !ok {
					errors += fmt.Sprintf("Error processing GET entity uri for resource %s, the URI template references a resource %s, but could not find it", key, resourceName)
				}
			}
		}

		if val.GetCollectionInfo != nil {
			template, err := uritemplate.New(val.GetCollectionInfo.Url)

			if err != nil {
				errors += fmt.Sprintf("Could not process GET collection uri for resource %s, error:%s\n", key, err)
			}

			for _, variable := range template.Varnames() {
				resourceName := strings.ReplaceAll(variable, "_", "-")
				if _, ok := resources[resourceName]; !ok {
					errors += fmt.Sprintf("Error processing GET collection uri for resource %s, the URI template references a resource %s, but could not find it", key, resourceName)
				}
			}
		}
	}

	// Verification

	if len(errors) > 0 {
		t.Fatalf("Errors occurred while validating URI Templates:\n%s", errors)
	}
}
