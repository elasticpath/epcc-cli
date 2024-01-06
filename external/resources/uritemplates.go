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
		resourceType := ConvertUriTemplateValueToType(varName)
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

	uri, err := template.Expand(values)

	if err != nil {
		return "", err
	}

	decodedUrlBytes, err := url.PathUnescape(uri)

	return decodedUrlBytes, err

}

func GenerateUrl(urlInfo *CrudEntityInfo, args []string, useAliases bool) (string, error) {
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
		resourceType := ConvertUriTemplateValueToType(varName)
		varType, ok := GetResourceByName(resourceType)
		if ok {
			attribute := "id"
			if override, ok := urlInfo.ParentResourceValueOverrides[resourceType]; ok {
				log.Tracef("url %s uses a type [%s] instead of id, so URL will be filled with this", urlInfo.Url, override)
				attribute = override
			}
			if useAliases {
				values[varName] = uritemplate.String(aliases.ResolveAliasValuesOrReturnIdentity(varType.JsonApiType, varType.AlternateJsonApiTypesForAliases, args[idx], attribute))
			} else {
				values[varName] = uritemplate.String(args[idx])
			}
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

		results = append(results, ConvertUriTemplateValueToType(value))
	}

	return results, nil
}

func GetSingularTypesOfVariablesNeeded(url string) ([]string, error) {
	var ret []string
	types, err := GetTypesOfVariablesNeeded(url)

	if err != nil {
		return nil, err
	}

	for _, t := range types {

		otherType, ok := GetResourceByName(t)

		if !ok {
			log.Warnf("Error processing resource, could not find type %s", t)
		}

		ret = append(ret, otherType.SingularName)
	}

	return ret, nil
}

func ConvertUriTemplateValueToType(value string) string {
	// URI templates must use _, so let's swap them for -
	return strings.ReplaceAll(value, "_", "-")
}
