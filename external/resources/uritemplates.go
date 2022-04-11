package resources

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/aliases"
	log "github.com/sirupsen/logrus"
	"github.com/yosida95/uritemplate/v3"
	"strings"
)

func GenerateUrl(resource Resource, url string, args []string) (string, error) {
	template, err := uritemplate.New(url)

	if err != nil {
		return "", fmt.Errorf("Could not generate URI template for URL: %w", err)
	}

	vars := template.Varnames()

	if len(vars) > len(args) {
		return "", fmt.Errorf("URI Template requires %d arguments, but only %d were passed", len(vars), len(args))
	}

	values := uritemplate.Values{}

	for idx, varName := range vars {
		resourceType := convertUriTemplateValueToType(varName)
		varType, ok := GetResourceByName(resourceType)
		if ok {
			values[varName] = uritemplate.String(aliases.ResolveAliasValuesOrReturnIdentity(varType.JsonApiType, args[idx]))
		} else {
			log.Warnf("Could not find a resource with type %s, aliases are probably broken", resourceType)
			values[varName] = uritemplate.String(args[idx])
		}

	}

	return template.Expand(values)
}

func GetNumberOfVariablesNeeded(url string) (int, error) {
	template, err := uritemplate.New(url)

	if err != nil {
		return 0, fmt.Errorf("Could not generate URI template for URL: %w", err)
	}

	return len(template.Varnames()), nil
}

func GetTypesOfVariablesNeeded(url string) ([]string, error) {

	results := make([]string, 0)

	template, err := uritemplate.New(url)

	if err != nil {
		return results, fmt.Errorf("Could not generate URI template for URL: %w", err)
	}

	for _, value := range template.Varnames() {

		results = append(results, convertUriTemplateValueToType(value))
	}

	return results, nil
}

func convertUriTemplateValueToType(value string) string {
	// URI templates must use _, so let's swap them for -
	return strings.ReplaceAll(value, "_", "-")
}
