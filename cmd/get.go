package cmd

import (
	"context"
	gojson "encoding/json"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/crud"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
)

const singularResourceRequest = 0
const collectionResourceRequest = 1

func NewGetCommand(parentCmd *cobra.Command) {

	overrides := &httpclient.HttpParameterOverrides{
		QueryParameters: nil,
		OverrideUrlPath: "",
	}

	var outputJq = ""

	var getCmd = &cobra.Command{
		Use:          "get",
		Short:        "Retrieves either a single or all resources",
		SilenceUsage: false,
	}

	for name, resource := range resources.GetPluralResources() {
		name := name
		resource := resource

		for i := 0; i < 2; i++ {
			i := i
			usageString := ""
			resourceName := ""
			completionVerb := 0
			usageGetType := ""
			var urlInfo *resources.CrudEntityInfo = nil

			switch i {
			case singularResourceRequest:
				if resource.GetCollectionInfo == nil {
					continue
				}

				resourceName = resource.PluralName
				usageString = resource.PluralName

				urlInfo = resource.GetCollectionInfo
				completionVerb = completion.GetAll
				usageGetType = "all (or a single page) of"

			case collectionResourceRequest:
				if resource.GetEntityInfo == nil {
					continue
				}
				usageString = resource.SingularName
				resourceName = resource.SingularName

				urlInfo = resource.GetEntityInfo
				completionVerb = completion.Get
				usageGetType = "a single"
			}

			resourceUrl := urlInfo.Url
			types, err := resources.GetTypesOfVariablesNeeded(resourceUrl)

			if err != nil {
				log.Warnf("Error processing resource %s, could not determine types from resource url %s", name, resourceUrl)
			}

			singularTypeNames := GetSingularTypeNames(types)

			exampleWithIds := fmt.Sprintf("  epcc get %s %s", resourceName, GetArgumentExampleWithIds(singularTypeNames))
			exampleWithAliases := fmt.Sprintf("  epcc get %s %s", resourceName, GetArgumentExampleWithAlias(singularTypeNames))

			parametersLongUsage := GetParameterUsageForTypes(singularTypeNames)

			usageString += GetParametersForTypes(singularTypeNames)

			examples := fmt.Sprintf("  # Retrieve %s %s\n%s\n  > GET %s\n\n", usageGetType, resourceName, exampleWithIds, FillUrlWithIds(urlInfo))

			if len(types) > 0 {
				examples += fmt.Sprintf("  # Retrieve %s %s using aliases \n%s\n  > GET %s\n\n", usageGetType, resourceName, exampleWithAliases, FillUrlWithIds(urlInfo))
			}

			queryParameters, _ := completion.Complete(completion.Request{
				Type:     completion.CompleteQueryParamKey,
				Resource: resource,
				Verb:     completionVerb,
			})

			for _, qp := range queryParameters {
				if qp == "" {
					continue
				}

				switch qp {
				case "page[limit]":
					examples += fmt.Sprintf("  # Retrieve %s %s with page[limit] = 25 and page[offset] = 500 \n%s %s %s %s %s \n > GET %s \n\n", usageGetType, resourceName, exampleWithAliases, qp, "25", "page[offset]", "500", FillUrlWithIds(urlInfo)+"?page[limit]=25&page[offset]=500")
					usageString += fmt.Sprintf(" [page[limit] N]")
				case "page[offset]":
					// No example
					usageString += fmt.Sprintf(" [page[offset] N]")
				case "sort":
					usageString += fmt.Sprintf(" [sort SORT]")
					sortKeys, _ := completion.Complete(completion.Request{
						Type:       completion.CompleteQueryParamValue,
						Resource:   resource,
						QueryParam: "sort",
						Verb:       completionVerb,
					})

					rand.Shuffle(len(sortKeys), func(i, j int) {
						sortKeys[i], sortKeys[j] = sortKeys[j], sortKeys[i]
					})

					for i, v := range sortKeys {
						if v[0] != '-' {
							examples += fmt.Sprintf("  # Retrieve %s %s sorted in ascending order of %s\n%s %s %s \n > GET %s\n\n", usageGetType, resourceName, v, exampleWithAliases, qp, v, FillUrlWithIds(urlInfo)+"?sort="+v)
						} else {
							examples += fmt.Sprintf("  # Retrieve %s %s sorted in descending order of %s\n%s %s -- %s\n > GET %s\n\n", usageGetType, resourceName, v, exampleWithAliases, qp, v, FillUrlWithIds(urlInfo)+"?sort="+v)
						}

						if i > 2 {
							// Only need three examples for sort
							break
						}
					}

				case "filter":
					usageString += fmt.Sprintf(" [filter FILTER]")
					attributeKeys, _ := completion.Complete(completion.Request{
						Type:       completion.CompleteAttributeKey,
						Resource:   resource,
						Attributes: map[string]int{},
						Verb:       completion.Create,
					})

					rand.Shuffle(len(attributeKeys), func(i, j int) {
						attributeKeys[i], attributeKeys[j] = attributeKeys[j], attributeKeys[i]
					})

					searchOps := []string{"eq", "like", "gt"}
					for i, v := range attributeKeys {
						examples += fmt.Sprintf(`  # Retrieve %s %s with filter %s(%s,"Hello World")
  %s %s '%s(%s,"Hello World")'
 > GET %s

`, usageGetType, resourceName, searchOps[i], v, exampleWithAliases, qp, searchOps[i], v, FillUrlWithIds(urlInfo)+fmt.Sprintf(`?filter=%s(%s,"Hello World")`, searchOps[i], v))

						if i >= 2 {
							// Only need three examples for sort
							break
						}
					}

				default:
					usageString += fmt.Sprintf(" [%s VALUE]", qp)
					examples += fmt.Sprintf("  # Retrieve %s %s with a(n) %s = %s\n%s %s %s \n > GET %s \n\n", usageGetType, resourceName, qp, "x", exampleWithAliases, qp, "x", FillUrlWithIds(urlInfo)+"?"+qp+"=x")
				}

			}

			newCmd := &cobra.Command{
				Use: usageString,
				// The replace all is a hack for the moment the URL could be made nicer
				Short: fmt.Sprintf("Calls %s", GetHelpResourceUrls(resourceUrl)),
				// The double "  " to " " is just a hack cause I was lazy

				Long: fmt.Sprintf(`Retrieves %s %s defined in a store/organization by calling %s.

%s
`, usageGetType, resourceName, GetHelpResourceUrls(resourceUrl), parametersLongUsage),
				Example: strings.ReplaceAll(strings.Trim(examples, "\n"), "  ", " "),
				Args:    GetArgsFunctionForResource(singularTypeNames),
				RunE: func(cmd *cobra.Command, args []string) error {

					body, err := getInternal(context.Background(), overrides, append([]string{resourceName}, args...))
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

					if resourceUrl == "" {
						return []string{}, cobra.ShellCompDirectiveNoFileComp
					}

					idCount, err := resources.GetNumberOfVariablesNeeded(resourceUrl)

					if err != nil {
						return []string{}, cobra.ShellCompDirectiveNoFileComp
					}

					if len(args) < idCount {
						// Must be for a resource completion
						types, err := resources.GetTypesOfVariablesNeeded(resourceUrl)

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

					} else if len(args) >= idCount { // Arg is after IDs
						if (len(args)-idCount)%2 == 0 { // This is a query param key
							return completion.Complete(completion.Request{
								Type:     completion.CompleteQueryParamKey,
								Resource: resource,
								Verb:     completionVerb,
							})

						} else {
							return completion.Complete(completion.Request{
								Type:       completion.CompleteQueryParamValue,
								Resource:   resource,
								Verb:       completionVerb,
								QueryParam: args[len(args)-1],
								ToComplete: toComplete,
							})

						}
					}

					return []string{}, cobra.ShellCompDirectiveNoFileComp
				},
			}

			getCmd.AddCommand(newCmd)
		}
	}

	getCmd.PersistentFlags().StringVar(&overrides.OverrideUrlPath, "override-url-path", "", "Override the URL that will be used for the Request")
	getCmd.PersistentFlags().StringSliceVarP(&overrides.QueryParameters, "query-parameters", "q", []string{}, "Pass in key=value an they will be added as query parameters")
	getCmd.PersistentFlags().StringVarP(&outputJq, "output-jq", "", "", "A jq expression, if set we will restrict output to only this")
	_ = getCmd.RegisterFlagCompletionFunc("output-jq", jqCompletionFunc)

	parentCmd.AddCommand(getCmd)
}

func getInternal(ctx context.Context, overrides *httpclient.HttpParameterOverrides, args []string) (string, error) {
	resp, err := getResource(ctx, overrides, args)

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
			json.PrintJson(string(body))
			return "", fmt.Errorf(resp.Status)
		}

		aliases.SaveAliasesForResources(string(body))

		return string(body), nil
	} else {
		return "", nil
	}
}

func getUrl(resource resources.Resource, args []string) (*resources.CrudEntityInfo, error) {

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

func getResource(ctx context.Context, overrides *httpclient.HttpParameterOverrides, args []string) (*http.Response, error) {
	crud.OutstandingRequestCounter.Add(1)
	defer crud.OutstandingRequestCounter.Done()

	// Find Resource
	resource, ok := resources.GetResourceByName(args[0])
	if !ok {
		return nil, fmt.Errorf("could not find resource %s", args[0])
	}

	var idCount int

	resourceUrlInfo, err2 := getUrl(resource, args)
	if err2 != nil {
		return nil, err2
	}

	idCount, err := resources.GetNumberOfVariablesNeeded(resourceUrlInfo.Url)

	if err != nil {
		return nil, err
	}

	// Replace ids with args in resourceURL
	resourceURL, err := resources.GenerateUrl(resourceUrlInfo, args[1:])

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

	for _, v := range overrides.QueryParameters {
		keyAndValue := strings.SplitN(v, "=", 2)
		if len(keyAndValue) != 2 {
			return nil, fmt.Errorf("Could not parse query parameter %v, all query parameters should be a key and value format", keyAndValue)
		}
		params.Add(keyAndValue[0], keyAndValue[1])
	}

	// Steve doesn't understand this logic check
	if (idCount-len(args)+1)%2 != 0 {
		resourceURL = resourceURL + url.QueryEscape(args[len(args)-1])
	}

	// Submit request
	resp, err := httpclient.DoRequest(ctx, "GET", resourceURL, params.Encode(), nil)

	if err != nil {
		return nil, fmt.Errorf("got error %s", err.Error())
	}

	return resp, nil
}
