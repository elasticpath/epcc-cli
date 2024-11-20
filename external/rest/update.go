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
	"net/url"
	"strings"
)

func UpdateInternal(ctx context.Context, overrides *httpclient.HttpParameterOverrides, skipAliases bool, args []string) (string, error) {
	shutdown.OutstandingOpCounter.Add(1)
	defer shutdown.OutstandingOpCounter.Done()

	// Find Resource
	resource, ok := resources.GetResourceByName(args[0])
	if !ok {
		return "", fmt.Errorf("could not find resource %s", args[0])
	}

	if resource.UpdateEntityInfo == nil {
		return "", fmt.Errorf("resource %s doesn't support UPDATE", args[0])
	}

	// Count ids in UpdateEntity
	resourceUrlInfo := resource.UpdateEntityInfo
	idCount, err := resources.GetNumberOfVariablesNeeded(resourceUrlInfo.Url)
	if err != nil {
		return "", err
	}

	// Replace ids with args in resourceURL
	resourceURL, err := resources.GenerateUrl(resourceUrlInfo, args[1:], true)
	if err != nil {
		return "", err
	}

	if overrides.OverrideUrlPath != "" {
		log.Warnf("Overriding URL Path from %s to %s", resourceURL, overrides.OverrideUrlPath)
		resourceURL = overrides.OverrideUrlPath
	}

	args = append(args, "type", resource.JsonApiType)
	// Create the body from remaining args
	body, err := json.ToJson(args[(idCount+1):], resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes, true)
	if err != nil {
		return "", err
	}

	params := url.Values{}

	for _, v := range overrides.QueryParameters {
		keyAndValue := strings.SplitN(v, "=", 2)
		if len(keyAndValue) != 2 {
			return "", fmt.Errorf("Could not parse query parameter %v, all query parameters should be a key and value format", keyAndValue)
		}
		params.Add(keyAndValue[0], keyAndValue[1])
	}

	// Submit request
	resp, err := httpclient.DoRequest(ctx, "PUT", resourceURL, params.Encode(), strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("got error %s", err.Error())
	} else if resp == nil {
		return "", fmt.Errorf("got nil response")
	}

	if resp.Body != nil {
		defer resp.Body.Close()

		// Print the body
		resBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		// Check if error response
		if resp.StatusCode >= 400 && resp.StatusCode <= 600 {
			json.PrintJson(string(resBody))
			return "", fmt.Errorf(resp.Status)
		}

		// 204 is no content, so we will skip it.
		if resp.StatusCode != 204 {
			if !skipAliases {
				aliases.SaveAliasesForResources(string(resBody))
			}
		}

		return string(resBody), nil
	} else {
		return "", nil
	}
}
