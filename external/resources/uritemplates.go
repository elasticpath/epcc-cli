package resources

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/id"
	log "github.com/sirupsen/logrus"
	"github.com/yosida95/uritemplate/v3"
	"net/url"
	"strings"
)

func GenerateUrlViaIdableAttributes(urlInfo *CrudEntityInfo, args []id.IdableAttributes) (string, error) {

	template, err := uritemplate.New(urlInfo.Url)

	if err != nil {
		return "", fmt.Errorf("could not generate URI template for URL: %w", err)
	}

	vars := template.Varnames()

	if len(vars) > len(args) {
		return "", fmt.Errorf("URI Template requires %d arguments, but only %d were passed", len(vars), len(args))
	}

	values := uritemplate.Values{}

	for idx, varName := range vars {
		resourceType := convertUriTemplateValueToType(varName)
		_, ok := GetResourceByName(resourceType)
		if ok {
			attribute := "id"

			if override, ok := urlInfo.ParentResourceValueOverrides[resourceType]; ok {
				log.Tracef("url %s uses a type [%s] instead of id, so URL will be filled with this", urlInfo.Url, override)
				attribute = override
			}

			value := ""
			if attribute == "id" {
				value = args[idx].Id
			}

			if attribute == "slug" {
				value = args[idx].Slug
			}

			if attribute == "sku" {
				value = args[idx].Sku
			}

			if value == "" {
				log.Warnf("Value for attribute %s is empty, url may not generate correctly", attribute)
			}

			values[varName] = uritemplate.String(value)

		} else {
			log.Warnf("Could not find a resource with type %s, aliases are probably broken", resourceType)
			values[varName] = uritemplate.String(args[idx].Id)
		}

	}

	return template.Expand(values)

}

func GenerateUrl(urlInfo *CrudEntityInfo, args []string) (string, error) {
	template, err := uritemplate.New(urlInfo.Url)

	if err != nil {
		return "", fmt.Errorf("could not generate URI template for URL: %w", err)
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
			attribute := "id"
			if override, ok := urlInfo.ParentResourceValueOverrides[resourceType]; ok {
				log.Tracef("url %s uses a type [%s] instead of id, so URL will be filled with this", urlInfo.Url, override)
				attribute = override
			}
			values[varName] = uritemplate.String(aliases.ResolveAliasValuesOrReturnIdentity(varType.JsonApiType, args[idx], attribute))
		} else {
			log.Warnf("Could not find a resource with type %s, aliases are probably broken", resourceType)
			values[varName] = uritemplate.String(args[idx])
		}

	}

	encodedUrl, err := template.Expand(values)
	if err != nil {
		return "", err
	}

	decodedUrlBytes, err := url.PathUnescape(encodedUrl)

	return decodedUrlBytes, err

}

func GetNumberOfVariablesNeeded(url string) (int, error) {
	template, err := uritemplate.New(url)

	if err != nil {
		return 0, fmt.Errorf("could not generate URI template for URL: %w", err)
	}

	return len(template.Varnames()), nil
}

func GetTypesOfVariablesNeeded(url string) ([]string, error) {

	results := make([]string, 0)

	template, err := uritemplate.New(url)

	if err != nil {
		return results, fmt.Errorf("could not generate URI template for URL: %w", err)
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
