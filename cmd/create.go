package cmd

import (
	"context"
	gojson "encoding/json"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/autofill"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/crud"
	"github.com/elasticpath/epcc-cli/external/encoding"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func NewCreateCommand(parentCmd *cobra.Command) {

	var autoFillOnCreate = false

	var outputJq = ""
	overrides := &httpclient.HttpParameterOverrides{
		QueryParameters: nil,
		OverrideUrlPath: "",
	}

	var create = &cobra.Command{
		Use:          "create",
		Short:        "Creates a resource",
		SilenceUsage: false,
	}

	for name, resource := range resources.GetPluralResources() {
		name := name
		resource := resource

		if resource.CreateEntityInfo == nil {
			continue
		}

		usageString := resource.SingularName
		resourceName := resource.SingularName
		resourceUrl := resource.CreateEntityInfo.Url
		//		completionVerb := completion.Create

		types, err := resources.GetTypesOfVariablesNeeded(resourceUrl)

		if err != nil {
			log.Warnf("Error processing resource %s, could not determine types from resource url %s", name, resourceUrl)
		}

		singularTypeNames := GetSingularTypeNames(types)
		usageString += GetParametersForTypes(singularTypeNames) + GetJsonKeyValuesForUsage(resource)

		exampleWithIds := fmt.Sprintf("  epcc create %s %s", resourceName, GetArgumentExampleWithIds(singularTypeNames))
		exampleWithAliases := fmt.Sprintf("  epcc create %s %s", resourceName, GetArgumentExampleWithAlias(singularTypeNames))

		parametersLongUsage := GetParameterUsageForTypes(types)

		baseJsonArgs := []string{}
		if !resource.NoWrapping {
			baseJsonArgs = append(baseJsonArgs, "type", resource.JsonApiType)
		}

		emptyJson, _ := json.ToJson(baseJsonArgs, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes)

		examples := GetJsonExample(fmt.Sprintf("# Create a %s", resource.SingularName), exampleWithIds, fmt.Sprintf("> POST %s", FillUrlWithIds(resource.CreateEntityInfo)), emptyJson)

		if len(types) > 0 {
			examples += GetJsonExample(fmt.Sprintf("# Create a %s using aliases", resource.SingularName), exampleWithIds, fmt.Sprintf("> POST %s", FillUrlWithIds(resource.CreateEntityInfo)), emptyJson)
		}

		if resource.CreateEntityInfo.ContentType != "multipart/form-data" {
			for k := range resource.Attributes {

				if k[0] == '^' {
					continue
				}

				results, _ := completion.Complete(completion.Request{
					Type:       completion.CompleteAttributeValue,
					Resource:   resource,
					Verb:       completion.Create,
					Attribute:  k,
					ToComplete: "",
				})

				arg := `"Hello World"`

				if len(results) > 0 {
					arg = results[0]
				}

				extendedArgs := append(baseJsonArgs, k, arg)

				// Don't try and use more than one key as some are mutually exclusive and the JSON will crash.
				// Resources that are heterogenous and can have array or object fields at some level (i.e., data[n].id and data.id) are examples
				jsonTxt, _ := json.ToJson(extendedArgs, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes)
				examples += GetJsonExample(fmt.Sprintf("# Create a %s passing in an argument", resourceName), fmt.Sprintf("%s %s %s", exampleWithAliases, k, arg), fmt.Sprintf("> POST %s", FillUrlWithIds(resource.CreateEntityInfo)), jsonTxt)

				autofilledData := autofill.GetJsonArrayForResource(&resource)

				extendedArgs = append(autofilledData, extendedArgs...)

				jsonTxt, _ = json.ToJson(extendedArgs, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes)
				examples += GetJsonExample(fmt.Sprintf("# Create a %s (using --auto-fill) and passing in an argument", resourceName), fmt.Sprintf("%s --auto-fill %s %s", exampleWithAliases, k, arg), fmt.Sprintf("> POST %s", FillUrlWithIds(resource.CreateEntityInfo)), jsonTxt)

				break
			}
		}

		argumentsBlurb := ""

		switch resource.CreateEntityInfo.ContentType {
		case "multipart/form-data":
			argumentsBlurb = "Key and values are passed in using multipart/form-data encoding\n\nDocumentation:\n  " + resource.CreateEntityInfo.Docs
		case "application/json", "":
			argumentsBlurb = fmt.Sprintf(`
Key and value pairs passed in will be converted to JSON with a jq like syntax.

The EPCC CLI will automatically determine appropriate wrapping

Basic Types:
key b => { "a": "b" }
key 1 => { "a": 1  }
key '"1"' => { "a": "1" }
key true => { "a": true }
key null => { "a": null }
key '"null"'' => { "a": "null" }



Documentation:
 %s
`, resource.CreateEntityInfo.Docs)
		default:
			argumentsBlurb = fmt.Sprintf("This resource uses %s encoding, which this help doesn't know how to help you with :) Submit a bug please.\nDocumentation:\n  %s", resource.CreateEntityInfo.ContentType, resource.CreateEntityInfo.Docs)
		}

		var createResourceCmd = &cobra.Command{
			Use:   usageString,
			Short: fmt.Sprintf("Calls %s", GetHelpResourceUrls(resourceUrl)),
			Long: fmt.Sprintf(`Creates a %s in a store/organization by calling %s.
%s
%s
`, resourceName, GetHelpResourceUrls(resourceUrl), parametersLongUsage, argumentsBlurb),
			Example: strings.ReplaceAll(strings.Trim(examples, "\n"), "  ", " "),
			Args:    GetArgsFunctionForResource(singularTypeNames),
			RunE: func(cmd *cobra.Command, args []string) error {
				body, err := createInternal(context.Background(), overrides, append([]string{name}, args...), autoFillOnCreate)

				if err != nil {
					return err
				}

				if outputJq != "" {
					output, err := json.RunJQOnStringWithArray(outputJq, body)

					if err != nil {
						return err
					}

					for _, outputLine := range output {
						outputJson, err := gojson.Marshal(outputLine)

						if err != nil {
							return err
						}

						err = json.PrintJson(string(outputJson))

						if err != nil {
							return err
						}
					}

					return nil
				}

				return json.PrintJson(body)
			},

			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				// Find Resource
				resource, ok := resources.GetResourceByName(resourceName)
				if ok {
					if resource.CreateEntityInfo != nil {
						resourceURL := resource.CreateEntityInfo.Url
						idCount, _ := resources.GetNumberOfVariablesNeeded(resourceURL)
						if len(args)-idCount >= 0 { // Arg is after IDs
							if (len(args)-idCount)%2 == 0 { // This is an attribute key
								usedAttributes := make(map[string]int)
								for i := idCount + 1; i < len(args); i = i + 2 {
									usedAttributes[args[i]] = 0
								}

								// I think this allows you to complete the current argument
								// This is necessary because if you are using something with a wildcard or regex
								// You won't see it in the attribute list, and therefore it won't be able to auto complete it.
								// I now think this does nothing.
								toComplete := strings.ReplaceAll(toComplete, "<ENTER>", "")
								if toComplete != "" {
									usedAttributes[toComplete] = 0
								}
								return completion.Complete(completion.Request{
									Type:       completion.CompleteAttributeKey,
									Resource:   resource,
									Attributes: usedAttributes,
									Verb:       completion.Create,
								})
							} else { // This is an attribute value
								return completion.Complete(completion.Request{
									Type:       completion.CompleteAttributeValue,
									Resource:   resource,
									Verb:       completion.Create,
									Attribute:  args[len(args)-1],
									ToComplete: toComplete,
								})
							}
						} else {
							// Arg is in IDS
							// Must be for a resource completion
							types, err := resources.GetTypesOfVariablesNeeded(resourceURL)

							if err != nil {
								return []string{}, cobra.ShellCompDirectiveNoFileComp
							}

							typeIdxNeeded := len(args) - 1

							if completionResource, ok := resources.GetResourceByName(types[typeIdxNeeded]); ok {
								return completion.Complete(completion.Request{
									Type:     completion.CompleteAlias,
									Resource: completionResource,
								})
							}
						}
					}
				}

				return []string{}, cobra.ShellCompDirectiveNoFileComp
			},
		}
		create.AddCommand(createResourceCmd)
	}

	create.PersistentFlags().StringVar(&overrides.OverrideUrlPath, "override-url-path", "", "Override the URL that will be used for the Request")
	create.PersistentFlags().BoolVarP(&autoFillOnCreate, "auto-fill", "", false, "Auto generate value for fields")
	create.PersistentFlags().StringSliceVarP(&overrides.QueryParameters, "query-parameters", "q", []string{}, "Pass in key=value an they will be added as query parameters")
	create.PersistentFlags().StringVarP(&outputJq, "output-jq", "", "", "A jq expression, if set we will restrict output to only this")

	_ = create.RegisterFlagCompletionFunc("output-jq", jqCompletionFunc)
	parentCmd.AddCommand(create)
}

func createInternal(ctx context.Context, overrides *httpclient.HttpParameterOverrides, args []string, autoFillOnCreate bool) (string, error) {
	crud.OutstandingRequestCounter.Add(1)
	defer crud.OutstandingRequestCounter.Done()

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
	resourceURL, err = resources.GenerateUrl(resource.CreateEntityInfo, args[1:])

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

		if !resource.NoWrapping {
			args = append(args, "type", resource.JsonApiType)
		}
		// Create the body from remaining args

		jsonArgs := args[(idCount + 1):]
		if autoFillOnCreate {
			autofilledData := autofill.GetJsonArrayForResource(&resource)

			jsonArgs = append(autofilledData, jsonArgs...)
		}

		body, err := json.ToJson(jsonArgs, resource.NoWrapping, resource.JsonApiFormat == "compliant", resource.Attributes)

		if err != nil {
			return "", err
		}

		// Submit request
		resp, err = httpclient.DoRequest(ctx, "POST", resourceURL, params.Encode(), strings.NewReader(body))

	}

	if err != nil {
		return "", fmt.Errorf("got error %s", err.Error())
	} else if resp == nil {
		return "", fmt.Errorf("got nil response")
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
			return "", fmt.Errorf(resp.Status)
		}

		// 204 is no content, so we will skip it.
		if resp.StatusCode != 204 {
			aliases.SaveAliasesForResources(string(resBody))
		}
		return string(resBody), nil
	} else {
		return "", nil
	}

}
