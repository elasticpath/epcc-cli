package rest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/autofill"
	"github.com/elasticpath/epcc-cli/external/encoding"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/elasticpath/epcc-cli/external/shutdown"
	log "github.com/sirupsen/logrus"
)

func CreateInternal(ctx context.Context, overrides *httpclient.HttpParameterOverrides, args []string, autoFillOnCreate bool, aliasName string, skipAliases bool, disableConstants bool, data string) (string, error) {
	shutdown.OutstandingOpCounter.Add(1)
	defer shutdown.OutstandingOpCounter.Done()

	// Find Resource
	resource, ok := resources.GetResourceByName(args[0])
	if !ok {
		return "", fmt.Errorf("could not find resource %s", args[0])
	}

	if resource.CreateEntityInfo == nil {
		return "", fmt.Errorf("resource %s doesn't support CREATE", args[0])
	}

	// Count ids in CreateEntity
	resourceURL := resource.CreateEntityInfo.Url

	idCount, err := resources.GetNumberOfVariablesNeeded(resourceURL)

	if err != nil {
		return "", err
	}

	// Replace ids with args in resourceURL
	resourceURL, err = resources.GenerateUrl(resource.CreateEntityInfo, args[1:], true)

	if overrides.OverrideUrlPath != "" {
		log.Warnf("Overriding URL Path from %s to %s", resourceURL, overrides.OverrideUrlPath)
		resourceURL = overrides.OverrideUrlPath
	}

	if err != nil {
		return "", err
	}

	var resp *http.Response = nil
	var resBody []byte

	if resource.CreateEntityInfo.ContentType == "multipart/form-data" {
		byteBuf, contentType, err := encoding.ToMultiPartEncoding(args[(idCount+1):], resource.NoWrapping, resource.JsonApiFormat == "complaint", resource.Attributes)
		if err != nil {
			return "", err
		}

		// Submit request
		resp, err = httpclient.DoFileRequest(ctx, resourceURL, byteBuf, contentType)

	} else {
		// Assume it's application/json

		params := url.Values{}

		for _, v := range overrides.QueryParameters {
			keyAndValue := strings.SplitN(v, "=", 2)
			if len(keyAndValue) != 2 {
				return "", fmt.Errorf("Could not parse query parameter %v, all query parameters should be a key and value format", keyAndValue)
			}
			params.Add(keyAndValue[0], keyAndValue[1])
		}

		var body string
		var err error

		if data != "" {
			// Use the provided data as the request body
			body = data
		} else {
			// Create the body from remaining args
			jsonArgs := args[(idCount + 1):]

			if !resource.NoWrapping && !disableConstants {
				jsonArgs = append([]string{"type", resource.JsonApiType}, jsonArgs...)
			}

			if autoFillOnCreate {
				autofilledData := autofill.GetJsonArrayForResource(&resource)
				jsonArgs = append(autofilledData, jsonArgs...)
			}

			body, err = json.ToJson(jsonArgs, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes, true, !disableConstants)
			if err != nil {
				return "", err
			}
		}

		// Submit request
		resp, err = httpclient.DoRequest(ctx, "POST", resourceURL, params.Encode(), strings.NewReader(body))
	}

	if err != nil {
		return "", fmt.Errorf("got error %s", err.Error())
	} else if resp == nil {
		return "", fmt.Errorf("got nil response with request: %s", resourceURL)
	}

	if resp.Body != nil {
		defer resp.Body.Close()

		// Print the body
		resBody, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		// Check if error response
		if resp.StatusCode >= 400 && resp.StatusCode <= 600 {
			json.PrintJson(string(resBody))
			return "", fmt.Errorf("%s", resp.Status)
		}

		// 204 is no content, so we will skip it.
		if resp.StatusCode != 204 {
			if !skipAliases {
				aliases.SaveAliasesForResources(string(resBody))
			}
		}

		if aliasName != "" {
			aliases.SetAliasForResource(string(resBody), aliasName)
		}

		return string(resBody), nil
	} else {
		return "", nil
	}
}
