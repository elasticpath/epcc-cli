package rest

import (
	"context"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/elasticpath/epcc-cli/external/shutdown"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func GetInternal(ctx context.Context, overrides *httpclient.HttpParameterOverrides, args []string, skipAliases bool) (string, error) {
	resp, err := GetResource(ctx, overrides, args)

	if err != nil {
		return "", err
	} else if resp == nil {
		return "", fmt.Errorf("got nil response")
	}

	if resp.Body != nil {
		defer resp.Body.Close()

		// Print the body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		// Check if error response
		if resp.StatusCode >= 400 && resp.StatusCode <= 600 {
			json.PrintJsonToStdout(string(body))
			return "", fmt.Errorf(resp.Status)
		}

		if !skipAliases {
			aliases.SaveAliasesForResources(string(body))
		}

		return string(body), nil
	} else {
		return "", nil
	}
}

func GetUrl(resource resources.Resource, args []string) (*resources.CrudEntityInfo, error) {

	if resource.GetCollectionInfo == nil && resource.GetEntityInfo == nil {
		return nil, fmt.Errorf("resource %s doesn't support GET", args[0])
	} else if resource.GetCollectionInfo != nil && resource.GetEntityInfo == nil {
		return resource.GetCollectionInfo, nil
	} else if resource.GetCollectionInfo == nil && resource.GetEntityInfo != nil {
		return resource.GetEntityInfo, nil
	} else {
		if _, ok := resources.GetPluralResources()[args[0]]; ok {
			return resource.GetCollectionInfo, nil
		} else {
			return resource.GetEntityInfo, nil
		}
	}
}

func GetResource(ctx context.Context, overrides *httpclient.HttpParameterOverrides, args []string) (*http.Response, error) {
	shutdown.OutstandingOpCounter.Add(1)
	defer shutdown.OutstandingOpCounter.Done()

	// Find Resource
	resource, ok := resources.GetResourceByName(args[0])
	if !ok {
		return nil, fmt.Errorf("could not find resource %s", args[0])
	}

	var idCount int

	resourceUrlInfo, err2 := GetUrl(resource, args)
	if err2 != nil {
		return nil, err2
	}

	idCount, err := resources.GetNumberOfVariablesNeeded(resourceUrlInfo.Url)

	if err != nil {
		return nil, err
	}

	// Replace ids with args in resourceURL
	resourceURL, err := resources.GenerateUrl(resourceUrlInfo, args[1:], true)

	if err != nil {
		return nil, err
	}

	if overrides.OverrideUrlPath != "" {
		log.Warnf("Overriding URL Path from %s to %s", resourceURL, overrides.OverrideUrlPath)
		resourceURL = overrides.OverrideUrlPath
	}

	// Add remaining args as query params
	params := url.Values{}
	for i := idCount + 1; i+1 < len(args); i = i + 2 {
		params.Add(args[i], args[i+1])
	}

	if (idCount-len(args)+1)%2 != 0 {
		log.Warnf("Extra argument at the end of the command %s", args[len(args)-1])
	}

	for _, v := range overrides.QueryParameters {
		keyAndValue := strings.SplitN(v, "=", 2)
		if len(keyAndValue) != 2 {
			return nil, fmt.Errorf("Could not parse query parameter %v, all query parameters should be a key and value format", keyAndValue)
		}
		params.Add(keyAndValue[0], keyAndValue[1])
	}

	// Submit request
	resp, err := httpclient.DoRequest(ctx, "GET", resourceURL, params.Encode(), nil)

	if err != nil {
		return nil, fmt.Errorf("got error %s", err.Error())
	}

	return resp, nil
}
