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

	var create = &cobra.Command{
		Use:   "create",
		Short: "Creates a resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("please specify a resource, epcc create [RESOURCE], see epcc create --help")
			} else {
				return fmt.Errorf("invalid resource [%s] specified, see all with epcc create --help", args[0])
			}
		},
	}

	for _, resource := range resources.GetPluralResources() {

		var autoFillOnCreate = false

		var noBodyPrint = false

		var outputJq = ""

		var setAlias = ""
		overrides := &httpclient.HttpParameterOverrides{
			QueryParameters: nil,
			OverrideUrlPath: "",
		}

		if resource.CreateEntityInfo == nil {
			continue
		}

		resource := resource
		resourceName := resource.SingularName

		var createResourceCmd = &cobra.Command{
			Use:     GetCreateUsageString(resource),
			Short:   GetCreateShort(resource),
			Long:    GetCreateLong(resource),
			Example: GetCreateExample(resource),
			Args:    GetArgFunctionForCreate(resource),
			RunE: func(cmd *cobra.Command, args []string) error {
				body, err := createInternal(context.Background(), overrides, append([]string{resourceName}, args...), autoFillOnCreate, setAlias)

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

				if noBodyPrint {
					return nil
				} else {
					return json.PrintJson(body)
				}

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
								for i := idCount; i < len(args); i = i + 2 {
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
									ToComplete: toComplete,
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

							typeIdxNeeded := len(args)

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

		createResourceCmd.PersistentFlags().StringVar(&overrides.OverrideUrlPath, "override-url-path", "", "Override the URL that will be used for the Request")
		createResourceCmd.PersistentFlags().BoolVarP(&autoFillOnCreate, "auto-fill", "", false, "Auto generate value for fields")
		createResourceCmd.PersistentFlags().BoolVarP(&noBodyPrint, "silent", "s", false, "Don't print the body on success")
		createResourceCmd.PersistentFlags().StringSliceVarP(&overrides.QueryParameters, "query-parameters", "q", []string{}, "Pass in key=value an they will be added as query parameters")
		createResourceCmd.PersistentFlags().StringVarP(&outputJq, "output-jq", "", "", "A jq expression, if set we will restrict output to only this")
		createResourceCmd.PersistentFlags().StringVarP(&setAlias, "save-as-alias", "", "", "A name to save the created resource as")

		_ = createResourceCmd.RegisterFlagCompletionFunc("output-jq", jqCompletionFunc)

		create.AddCommand(createResourceCmd)
	}

	parentCmd.AddCommand(create)
}

func createInternal(ctx context.Context, overrides *httpclient.HttpParameterOverrides, args []string, autoFillOnCreate bool, aliasName string) (string, error) {
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

		if aliasName != "" {
			aliases.SetAliasForResource(string(resBody), aliasName)
		}
		return string(resBody), nil
	} else {
		return "", nil
	}

}
