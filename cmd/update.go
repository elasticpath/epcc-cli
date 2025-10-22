package cmd

import (
	"context"
	gojson "encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/profiles"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/elasticpath/epcc-cli/external/rest"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type ResourceCache struct {
	Key       []string  `json:"key"`
	Json      string    `json:"json"`
	Error     bool      `json:"error"`
	ExpiresAt time.Time `json:"expires_at"`
}

func NewUpdateCommand(parentCmd *cobra.Command) func() {
	overrides := &httpclient.HttpParameterOverrides{
		QueryParameters: nil,
		OverrideUrlPath: "",
	}

	// Ensure that any new options here are added to the resetFunc
	var outputJq = ""
	var compactOutput = false
	var noBodyPrint = false
	var ifAliasExists = ""
	var ifAliasDoesNotExist = ""
	var skipAliases = false
	var repeat uint32 = 1
	var repeatDelay uint32 = 100
	var ignoreErrors = false
	var logOnSuccess = ""
	var logOnFailure = ""
	var disableConstants = false
	var data = ""

	resetFunc := func() {
		overrides.QueryParameters = nil
		overrides.OverrideUrlPath = ""
		outputJq = ""
		compactOutput = false
		noBodyPrint = false
		ifAliasExists = ""
		ifAliasDoesNotExist = ""
		skipAliases = false
		repeat = 1
		repeatDelay = 100
		ignoreErrors = false
		logOnSuccess = ""
		logOnFailure = ""
		disableConstants = false
		data = ""
	}

	var updateCmd = &cobra.Command{
		Use:          "update",
		Short:        "Updates a resource",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("please specify a resource, epcc update [RESOURCE], see epcc update --help")
			} else {
				return fmt.Errorf("invalid resource [%s] specified, see all with epcc update --help", args[0])
			}
		},
	}

	e := config.GetEnv()
	hiddenResources := map[string]struct{}{}
	for _, v := range e.EPCC_CLI_DISABLE_RESOURCES {
		hiddenResources[v] = struct{}{}
	}

	for _, resource := range resources.GetPluralResources() {
		resource := resource
		resourceName := resource.SingularName
		if resource.UpdateEntityInfo == nil {
			continue
		}

		if _, ok := hiddenResources[resource.SingularName]; ok {
			log.Tracef("Hiding resource %s", resource.SingularName)
			continue
		}

		if _, ok := hiddenResources[resource.PluralName]; ok {
			log.Tracef("Hiding resource %s", resource.SingularName)
			continue
		}

		var updateResourceCmd = &cobra.Command{
			Use:     GetUpdateUsage(resource),
			Short:   GetUpdateShort(resource),
			Long:    GetUpdateLong(resource),
			Example: GetUpdateExample(resource),
			Args:    GetArgFunctionForUpdate(resource),
			RunE: func(cmd *cobra.Command, args []string) error {
				c := func(cmd *cobra.Command, args []string) error {
					if ifAliasExists != "" {
						aliasId := aliases.ResolveAliasValuesOrReturnIdentity(resource.JsonApiType, resource.AlternateJsonApiTypesForAliases, ifAliasExists, "id")

						if aliasId == ifAliasExists {
							// If the aliasId is the same as requested, it means an alias did not exist.
							log.Infof("Alias [%s] does not exist, not continuing run", ifAliasExists)
							return nil
						}
					}

					if ifAliasDoesNotExist != "" {
						aliasId := aliases.ResolveAliasValuesOrReturnIdentity(resource.JsonApiType, resource.AlternateJsonApiTypesForAliases, ifAliasDoesNotExist, "id")

						if aliasId != ifAliasDoesNotExist {
							// If the aliasId is different from the request then it does exist.
							log.Infof("Alias [%s] does exist (value: %s), not continuing run", ifAliasDoesNotExist, aliasId)
							return nil
						}
					}

					body, err := rest.UpdateInternal(context.Background(), overrides, skipAliases, disableConstants, append([]string{resourceName}, args...), data)

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
						if compactOutput {
							body, err = json.Compact(body)

							if err != nil {
								return err
							}
						}

						return json.PrintJson(body)
					}
				}

				res := repeater(c, repeat, repeatDelay, cmd, args, ignoreErrors)

				if res != nil {
					if logOnFailure != "" {
						log.Errorf("%s", logOnFailure)
					}
				} else {
					if logOnSuccess != "" {
						log.Infof("%s", logOnSuccess)
					}
				}

				return res
			},

			ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				// Find Resource
				resourceURL := resource.UpdateEntityInfo.Url
				idCount, _ := resources.GetNumberOfVariablesNeeded(resourceURL)

				if len(args)-idCount >= 0 {
					// If the arg is after the id, then we are completing an attribute key.
					// Let's secretly fetch the current state of the resource so we can have better completion.
					ids := []string{}
					for id := 0; id < idCount; id++ {
						ids = append(ids, args[id])
					}

					ignoreConditionalChecks := false

					result, err := GetCurrentResourceState(resourceName, ids, overrides)

					existingAttributes := map[string]string{}
					if err != nil {
						log.Warnf("Could not get current state of resource, %v", err)
						ignoreConditionalChecks = true
					}

					if !ignoreConditionalChecks {
						v, err := json.FromJsonToMap(result)

						if err != nil {
							log.Warnf("Could not convert state to map, %v", err)
							ignoreConditionalChecks = true
						} else {
							existingAttributes = v
						}
					}

					// Arg is after IDs
					if (len(args)-idCount)%2 == 0 { // This is an attribute key
						usedAttributes := make(map[string]string)
						for i := idCount; i < len(args); i = i + 2 {
							if i+1 <= len(args) {
								usedAttributes[args[i]] = args[i+1]
							} else {
								usedAttributes[args[i]] = ""
							}
						}

						return completion.Complete(completion.Request{
							Type:                       completion.CompleteAttributeKey,
							Resource:                   resource,
							Attributes:                 usedAttributes,
							ExistingResourceAttributes: existingAttributes,
							SkipWhenChecksAndAddAll:    true,
							Verb:                       completion.Update,
						})
					} else { // This is an attribute value
						return completion.Complete(completion.Request{
							Type:           completion.CompleteAttributeValue,
							Resource:       resource,
							Verb:           completion.Update,
							Attribute:      args[len(args)-1],
							ToComplete:     toComplete,
							AllowTemplates: true,
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
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			},
		}

		updateCmd.AddCommand(updateResourceCmd)
	}

	updateCmd.PersistentFlags().StringVar(&overrides.OverrideUrlPath, "override-url-path", "", "Override the URL that will be used for the Request")
	updateCmd.PersistentFlags().StringSliceVarP(&overrides.QueryParameters, "query-parameters", "q", []string{}, "Pass in key=value an they will be added as query parameters")
	updateCmd.PersistentFlags().BoolVarP(&noBodyPrint, "silent", "s", false, "Don't print the body on success")
	updateCmd.PersistentFlags().StringVarP(&outputJq, "output-jq", "", "", "A jq expression, if set we will restrict output to only this")
	updateCmd.PersistentFlags().BoolVarP(&compactOutput, "compact", "", false, "Hides some of the boiler plate keys and empty fields, etc...")
	updateCmd.PersistentFlags().BoolVarP(&ignoreErrors, "ignore-errors", "", false, "Don't return non zero on an error")
	updateCmd.PersistentFlags().StringVarP(&ifAliasExists, "if-alias-exists", "", "", "If the alias exists we will run this command, otherwise exit with no error")
	updateCmd.PersistentFlags().StringVarP(&ifAliasDoesNotExist, "if-alias-does-not-exist", "", "", "If the alias does not exist we will run this command, otherwise exit with no error")
	updateCmd.MarkFlagsMutuallyExclusive("if-alias-exists", "if-alias-does-not-exist")
	updateCmd.PersistentFlags().BoolVarP(&skipAliases, "skip-alias-processing", "", false, "if set, we don't process the response for aliases")
	_ = updateCmd.RegisterFlagCompletionFunc("output-jq", jqCompletionFunc)
	updateCmd.PersistentFlags().Uint32VarP(&repeat, "repeat", "", 1, "Number of times to repeat the command")
	updateCmd.PersistentFlags().Uint32VarP(&repeatDelay, "repeat-delay", "", 100, "Delay (in ms) between repeats")

	updateCmd.PersistentFlags().BoolVarP(&disableConstants, "no-auto-constants", "", false, "Disable setting of known constant values in the request body (e.g., `type`)")
	updateCmd.PersistentFlags().StringVarP(&logOnSuccess, "log-on-success", "", "", "Output the following message as an info if the result is successful")
	updateCmd.PersistentFlags().StringVarP(&logOnFailure, "log-on-failure", "", "", "Output the following message as an error if the result fails")
	updateCmd.PersistentFlags().StringVarP(&data, "data", "d", "", "Raw JSON data to use as the request body. If provided, positional arguments will be ignored.")

	parentCmd.AddCommand(updateCmd)

	return resetFunc

}

func GetCurrentResourceState(resourceName string, ids []string, overrides *httpclient.HttpParameterOverrides) (string, error) {

	cacheFile := filepath.Clean(profiles.GetProfileDataDirectory() + "/update_completion_cache.json")

	myKey := append([]string{resourceName}, ids...)

	f, err := os.ReadFile(cacheFile)

	var cache ResourceCache

	if err == nil {
		err = gojson.Unmarshal(f, &cache)
		if err == nil {

			if slices.Equal(cache.Key, myKey) && cache.ExpiresAt.Unix() > time.Now().Unix() {
				if cache.Error {
					return "", errors.New("cache contains negative result")
				} else {
					return cache.Json, nil
				}

			}
		} else {
			log.Errorf("Could not unmarshal JSON cache for %s, error: %v", cacheFile, err)
		}
	} else {
		if errors.Is(err, os.ErrNotExist) {
			log.Infof("Could not read cache file %s, error: %v", cacheFile, err)
		} else {
			log.Warnf("Could not read cache file %s, error: %v", cacheFile, err)
		}
	}

	result, err := rest.GetInternal(context.Background(), overrides, append([]string{resourceName}, ids...), false, true)

	if err != nil {
		cache = ResourceCache{
			Key:       myKey,
			Json:      "",
			Error:     true,
			ExpiresAt: time.Now().Add(time.Second * 5),
		}
	} else {
		cache = ResourceCache{
			Key:       myKey,
			Json:      result,
			Error:     false,
			ExpiresAt: time.Now().Add(time.Second * 30),
		}
	}

	v, err := gojson.Marshal(cache)

	if err != nil {
		log.Warnf("Failed to convert cache to json: %v", err)
	} else {
		err = os.WriteFile(cacheFile, v, 0600)

		if err != nil {
			log.Warnf("Failed to save cache to file %s, error: %v", cacheFile, err)
		}
	}

	return result, err
}
