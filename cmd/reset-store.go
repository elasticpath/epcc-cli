package cmd

import (
	"context"
	gojson "encoding/json"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

var DeleteApplicationKeys = true

var ResetStore = &cobra.Command{
	Use:   "reset-store [STORE_ID]",
	Short: "Resets a store to it's initial state on a \"best effort\" basis.",
	Long:  "This command resets a store to it's initial state. There are some limitations to this as for instance orders cannot be deleted, nor can audit entries.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		overrides := &httpclient.HttpParameterOverrides{
			QueryParameters: nil,
			OverrideUrlPath: "",
		}

		storeId, err := getStoreId(ctx, args)
		if err != nil {
			return fmt.Errorf("could not determine store id: %w", err)
		}

		// We secretly support regexes for this command, it is undocumented however
		// internal users and power users can just pass .* to reset any store.
		rx, err := regexp.Compile("^" + args[0] + "$")

		if err != nil {
			if storeId != args[0] {
				return fmt.Errorf("You are trying to reset store id '%s', but you passed '%s' to this command", storeId, args[0])
			}
		} else {
			if !rx.MatchString(storeId) {
				return fmt.Errorf("You are trying to reset store id '%s', but you passed '%s' to this command which doesn't match", storeId, args[0])
			}
		}

		errors := make([]string, 0)

		// In theory we could topo-sort all the resources and determine dependencies.
		// We would also need locking to go faster.

		// Get customer and account authentication settings to populate the aliases
		_, err = getInternal(ctx, overrides, []string{"customer-authentication-settings"})

		if err != nil {
			errors = append(errors, err.Error())
		}

		_, err = getInternal(ctx, overrides, []string{"account-authentication-settings"})

		if err != nil {
			errors = append(errors, err.Error())
		}

		_, err = getInternal(ctx, overrides, []string{"merchant-realm-mappings"})

		if err != nil {
			errors = append(errors, err.Error())
		}

		_, err = getInternal(ctx, overrides, []string{"authentication-realms"})

		if err != nil {
			errors = append(errors, err.Error())
		}

		err, resetUndeletableResourcesErrors := resetResourcesUndeletableResources(ctx, overrides)

		if err != nil {
			return err
		}

		errors = append(errors, resetUndeletableResourcesErrors...)

		resourceNames := resources.GetPluralResourceNames()
		sort.Strings(resourceNames)
		err, deleteAllResourceDataErrors := deleteAllResourceData(resourceNames)
		if err != nil {
			return err
		}

		errors = append(errors, deleteAllResourceDataErrors...)

		// TODO core flows hack

		if len(errors) > 0 {
			log.Warnf("The following errors occurred while deleting all data: \n\t%s", strings.Join(errors, "\n\t"))
		}

		return nil

	},
}

func getStoreId(ctx context.Context, args []string) (string, error) {
	resource, ok := resources.GetResourceByName("settings")

	if !ok {
		return "", fmt.Errorf("could not find resource %s, we need it to determine the store id.", args[0])
	}

	resourceURL, err := resources.GenerateUrl(resource.GetCollectionInfo, make([]string, 0))

	if err != nil {
		return "", err
	}

	params := url.Values{}

	resp, err := httpclient.DoRequest(ctx, "GET", resourceURL, params.Encode(), nil)

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	var jsonStruct = map[string]interface{}{}
	err = gojson.Unmarshal(body, &jsonStruct)
	if err != nil {
		return "", err
	}

	storeIdInterface, err := json.RunJQ(".data.id", jsonStruct)

	if err != nil {
		return "", err
	}

	storeId, ok := storeIdInterface.(string)

	if !ok {
		return "", fmt.Errorf("Could not retrieve store id, could not cast result to string %T => %v", storeIdInterface, storeIdInterface)
	}
	return storeId, nil
}

func resetResourcesUndeletableResources(ctx context.Context, overrides *httpclient.HttpParameterOverrides) (error, []string) {

	resetCmds := [][]string{
		{"payment-gateway-adyen", "merchant_account", "", "username", "", "password", "", "enabled", "false", "test", "false"},
		{"payment-gateway-authorize-net", "login", "", "password", "", "enabled", "false", "test", "false"},
		{"payment-gateway-braintree", "merchant_id", "", "public_key", "", "environment", "", "enabled", "false"},
		{"payment-gateway-cardconnect", "merchant_id", "", "username", "", "password", "", "enabled", "false", "test", "false"},
		{"payment-gateway-cybersource", "login", "", "password", "", "enabled", "false", "test", "false"},
		{"payment-gateway-paypal-express-checkout", "payer_id", "", "enabled", "false", "test", "false"},
		{"payment-gateway-manual", "enabled", "false"},
		{"payment-gateway-stripe", "login", "", "enabled", "false"},
		{"payment-gateway-stripe-connect", "stripe_account", "", "enabled", "false", "test", "false"},
		{"payment-gateway-stripe-payment-intents", "login", "", "enabled", "false"},
		{"payment-gateway-elastic-path-payments-stripe", "stripe_account", "", "enabled", "false", "test", "false"},
		{"settings", "page_length", "25", "list_child_products", "true", "additional_languages", "[]", "calculation_method", "line", "address_mandatory_fields[0]", "first_name", "address_mandatory_fields[1]", "last_name", "address_mandatory_fields[2]", "line_1", "address_mandatory_fields[3]", "county", "address_mandatory_fields[4]", "postcode", "address_mandatory_fields[5]", "country"},
		// We can only use an alias for the customer authentication settings, MRM doesn't use a relationship, and Account management uses the wrong type.
		{"authentication-realm", "last_read=array[0]", "redirect_uris", "[]", "duplicate_email_policy", "allowed"},
		{"authentication-realm", "last_read=array[1]", "redirect_uris", "[]", "duplicate_email_policy", "allowed"},
		{"authentication-realm", "last_read=array[2]", "redirect_uris", "[]", "duplicate_email_policy", "allowed"},
		{"authentication-realm", "related_authentication-realm_for_customer-authentication-settings_last_read=entity", "name", "Buyer Organization"},
		{"authentication-realm", "related_authentication_realm_for_account_authentication_settings_last_read=entity", "name", "Account Management Realm"},
		{"log-ttl-settings", "days", "356"},
	}

	errors := make([]string, 0)

	for _, resetCmd := range resetCmds {
		body, err := updateInternal(ctx, overrides, resetCmd)

		if err != nil {
			errors = append(errors, fmt.Errorf("error resetting  %s: %v", resetCmd[0], err).Error())
		}

		err = json.PrintJson(body)

		if err != nil {
			errors = append(errors, fmt.Errorf("error resetting  %s: %v", resetCmd[0], err).Error())
		}
	}

	return nil, errors
}

func deleteAllResourceData(resourceNames []string) (error, []string) {
	noGetCollectionEndpoint := make([]string, 0)
	noDeleteEndpoint := make([]string, 0)
	ignoredEndpoints := make([]string, 0)

	maxDepth := 0
	errors := make([]string, 0)

	for _, resourceName := range resourceNames {
		resource, ok := resources.GetResourceByName(resourceName)

		if !ok {
			return fmt.Errorf("could not retrieve resource '%s'", resourceName), errors
		}

		if resource.GetCollectionInfo == nil {
			if !resource.SuppressResetWarning {
				if resource.CreateEntityInfo != nil || resource.UpdateEntityInfo != nil || resource.DeleteEntityInfo != nil {
					// If we can't mutate an entity, then lets assume that we don't need to reset it.
					noGetCollectionEndpoint = append(noGetCollectionEndpoint, resourceName)
				}

			}
			continue
		}

		if resource.DeleteEntityInfo == nil {
			if !resource.SuppressResetWarning {
				if resource.CreateEntityInfo != nil || resource.UpdateEntityInfo != nil {
					// If we can't mutate an entity, then lets assume that we don't need to reset it.
					noDeleteEndpoint = append(noDeleteEndpoint, resourceName)
				}
			}
			continue
		}

		if !DeleteApplicationKeys && resource.PluralName == "application-keys" {
			ignoredEndpoints = append(ignoredEndpoints, "application-keys")
		}

		myDepth, err := resources.GetNumberOfVariablesNeeded(resource.GetCollectionInfo.Url)

		if err != nil {
			return err, errors
		}

		if maxDepth < myDepth {
			maxDepth = myDepth
		}
	}

	log.Infof("Maximum depth of any resource is %d", maxDepth)

	for depth := maxDepth; depth >= 0; depth -= 1 {
		log.Infof("Processing all resources with depth %d", depth)
		for _, resourceName := range resourceNames {
			resource, ok := resources.GetResourceByName(resourceName)

			if !ok {
				return fmt.Errorf("could not retrieve resource '%s'", resourceName), errors
			}

			if resource.GetCollectionInfo == nil {
				continue
			}

			if resource.DeleteEntityInfo == nil {
				continue
			}

			if resource.PluralName == "application-keys" && !DeleteApplicationKeys {
				continue
			}

			myDepth, err := resources.GetNumberOfVariablesNeeded(resource.GetCollectionInfo.Url)

			if err != nil {
				return err, errors
			}

			if myDepth == depth {
				log.Infof("Processing resource %s", resourceName)
				err := deleteAllInternal(context.Background(), []string{resourceName})

				if err != nil {
					errors = append(errors, fmt.Errorf("error while deleting %s: %w", resourceName, err).Error())
				}
			}

		}
	}

	sort.Strings(noGetCollectionEndpoint)
	sort.Strings(noDeleteEndpoint)
	sort.Strings(ignoredEndpoints)

	log.Infof("The following %d resources were not deleted because we have no way to get a collection: %s", len(noGetCollectionEndpoint), strings.Join(noGetCollectionEndpoint, ", "))
	log.Infof("The following %d resources were not deleted because we have no way to delete an element: %s", len(noDeleteEndpoint), strings.Join(noDeleteEndpoint, ", "))
	log.Infof("The following %d resources were not deleted because they were ignored: %s", len(ignoredEndpoints), strings.Join(ignoredEndpoints, ", "))
	return nil, errors
}
