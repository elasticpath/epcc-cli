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

func DeleteResource(ctx context.Context, overrides *httpclient.HttpParameterOverrides, args []string) (*http.Response, error) {
	// Find Resource
	resource, ok := resources.GetResourceByName(args[0])
	if !ok {
		return nil, fmt.Errorf("could not find resource %s", args[0])
	}

	if resource.DeleteEntityInfo == nil {
		return nil, fmt.Errorf("resource %s doesn't support DELETE", args[0])
	}

	// Replace ids with args in resourceURL
	resourceURL, err := resources.GenerateUrl(resource.DeleteEntityInfo, args[1:], true)

	if err != nil {
		return nil, err
	}

	if overrides.OverrideUrlPath != "" {
		log.Warnf("Overriding URL Path from %s to %s", resourceURL, overrides.OverrideUrlPath)
		resourceURL = overrides.OverrideUrlPath
	}

	params := url.Values{}

	for _, v := range overrides.QueryParameters {
		keyAndValue := strings.SplitN(v, "=", 2)
		if len(keyAndValue) != 2 {
			return nil, fmt.Errorf("Could not parse query parameter %v, all query parameters should be a key and value format", keyAndValue)
		}
		params.Add(keyAndValue[0], keyAndValue[1])
	}

	idCount, err := resources.GetNumberOfVariablesNeeded(resource.DeleteEntityInfo.Url)

	if !resource.NoWrapping {
		args = append(args, "type", resource.JsonApiType)
	}
	// Create the body from remaining args

	jsonArgs := args[(idCount + 1):]

	var payload io.Reader = nil
	if len(jsonArgs) > 0 {
		body, err := json.ToJson(jsonArgs, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes, true)

		if err != nil {
			return nil, err
		}

		payload = strings.NewReader(body)
	}

	// Submit request
	resp, err := httpclient.DoRequest(ctx, "DELETE", resourceURL, params.Encode(), payload)
	if err != nil {
		return nil, fmt.Errorf("got error %s", err.Error())
	}

	return resp, nil
}

func DeleteInternal(ctx context.Context, overrides *httpclient.HttpParameterOverrides, allow404 bool, args []string) (string, error) {
	shutdown.OutstandingOpCounter.Add(1)
	defer shutdown.OutstandingOpCounter.Done()

	resource, ok := resources.GetResourceByName(args[0])
	if !ok {
		return "", fmt.Errorf("could not find resource %s", args[0])
	}

	resp, err := DeleteResource(ctx, overrides, args)
	if err != nil {
		return "", err
	}

	if resp == nil {
		return "", fmt.Errorf("got nil response")
	}

	if resp.StatusCode < 400 {
		idToDelete := aliases.ResolveAliasValuesOrReturnIdentity(resource.JsonApiType, resource.AlternateJsonApiTypesForAliases, args[len(args)-1], "id")
		aliases.DeleteAliasesById(idToDelete, resource.JsonApiType)
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
			if resp.StatusCode != 404 || !allow404 {
				return string(body), fmt.Errorf(resp.Status)
			}
		}

		return string(body), nil
	} else {
		return "", nil
	}

}
