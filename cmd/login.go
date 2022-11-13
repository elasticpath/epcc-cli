package cmd

import (
	gojson "encoding/json"
	"fmt"
	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/authentication"
	"github.com/elasticpath/epcc-cli/external/browser"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/url"
	"time"
)

const (
	API          = "api"
	ClientId     = "client_id"
	ClientSecret = "client_secret"
)

var loginCmd = &cobra.Command{
	Use:          "login",
	Short:        "Login to the API via client_credentials, implicit, customer or account management tokens.",
	SilenceUsage: false,
}

var loginInfo = &cobra.Command{
	Use:     "status",
	Short:   "Check the current (local) status of our authentication with the API",
	Aliases: []string{"info"},
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		apiTokenResponse := authentication.GetApiToken()

		if config.Envs.EPCC_BETA_API_FEATURES == "" {
			log.Infof("We have no configured API endpoint, will use default endpoint")
		} else {
			log.Infof("We are currently using API endpoint: %s", config.Envs.EPCC_API_BASE_URL)
		}

		if apiTokenResponse != nil {
			if time.Now().Unix() > apiTokenResponse.Expires {
				log.Infof("We are logged into the API with a %s type, but the token expired at: %s ", apiTokenResponse.Identifier, time.Unix(apiTokenResponse.Expires, 0).Format(time.RFC1123Z))
			} else {
				log.Infof("We are logged into the API with a %s type, the token expires at: %s", apiTokenResponse.Identifier, time.Unix(apiTokenResponse.Expires, 0).Format(time.RFC1123Z))
			}
		} else {
			log.Infof("We are *NOT* logged into the API")
		}

		customerTokenResponse := authentication.GetCustomerToken()
		if customerTokenResponse != nil {
			if time.Now().Unix() > customerTokenResponse.Data.Expires {
				log.Infof("We are using a customer token for customer %s <%s> (id=%s), but the token expired at: %s ", customerTokenResponse.AdditionalInfo.CustomerName, customerTokenResponse.AdditionalInfo.CustomerEmail, customerTokenResponse.Data.CustomerId, time.Unix(customerTokenResponse.Data.Expires, 0).Format(time.RFC1123Z))
			} else {
				log.Infof("We are using a customer token for customer %s <%s> (id=%s), the token expires at: %s", customerTokenResponse.AdditionalInfo.CustomerName, customerTokenResponse.AdditionalInfo.CustomerEmail, customerTokenResponse.Data.CustomerId, time.Unix(customerTokenResponse.Data.Expires, 0).Format(time.RFC1123Z))
			}

			if apiTokenResponse != nil && apiTokenResponse.Identifier == "client_credentials" {
				log.Warnf("You are current logged in with client_credentials and the customer token. Mixing client_credentials and customer token can lead to unintended results.")
			}
		}

		if authentication.IsAutoLoginEnabled() {
			if config.Envs.EPCC_CLIENT_SECRET != "" {
				log.Infof("Auto login is enabled and we will (attempt to) login with client_credentials")
			} else {
				log.Infof("Auto login is enabled and we will (attempt to) login with implicit, as no client_secret is available")
			}
		} else {

			if apiTokenResponse == nil {
				log.Warnf("Auto login is disabled, and we are not logged in. Most API calls will fail.")
			} else {
				log.Info("Auto login is disabled")
			}

		}

		log.Infof("All tokens are stored in %s", authentication.GetAuthenticationCacheDirectory())

		return nil
	},
}

var loginDocs = &cobra.Command{
	Use:       "docs {client_credentials|implicit|customer|account-member|permissions}",
	Short:     "Load documentation about authentication in the API",
	Args:      cobra.OnlyValidArgs,
	ValidArgs: []string{"client_credentials", "implicit", "customer", "customer-token", "account-member", "account-management-authentication-token", "account", "permissions"},
	RunE: func(cmd *cobra.Command, args []string) error {

		switch len(args) {
		case 0:
			return browser.OpenUrl("https://documentation.elasticpath.com/commerce-cloud/docs/api/basics/authentication/index.html")
		case 1:
			switch args[0] {
			case "client_credentials":
				return browser.OpenUrl("https://documentation.elasticpath.com/commerce-cloud/docs/api/basics/authentication/client-credential-token.html")
			case "implicit":
				return browser.OpenUrl("https://documentation.elasticpath.com/commerce-cloud/docs/api/basics/authentication/implicit-token.html")
			case "customer", "customer-token":
				return browser.OpenUrl("https://documentation.elasticpath.com/commerce-cloud/docs/api/basics/authentication/customer-token.html")
			case "account-member", "account", "account-management-authentication-token":
				return browser.OpenUrl("https://documentation.elasticpath.com/commerce-cloud/docs/api/basics/authentication/account-management-authentication-token.html")
			case "permissions":
				return browser.OpenUrl("https://documentation.elasticpath.com/commerce-cloud/docs/api/basics/authentication/permissions.html")
			default:
				panic("The valid args should have prevented this from happening")
			}
		default:
			panic("The valid args should have prevented this from happening")

		}

		return nil
	},
}

var loginClientCredentials = &cobra.Command{
	Use:   "client_credentials ([client_id <CLIENT_ID> client_secret <CLIENT_SECRET>])",
	Short: "Login via client credentials",
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 || (args[0] != "client_id" && args[0] != "client_secret") {
			return completion.Complete(completion.Request{
				Type: completion.CompleteLoginClientID + completion.CompleteLoginClientSecret,
			})
		} else if len(args) == 2 && (args[0] == "client_id") {
			return completion.Complete(completion.Request{
				Type: completion.CompleteLoginClientSecret,
			})
		} else if len(args) == 2 && args[0] == "client_secret" {
			return completion.Complete(completion.Request{
				Type: completion.CompleteLoginClientID,
			})
		} else {
			return completion.Complete(completion.Request{
				Type: 0,
			})
		}
	},

	RunE: func(cmd *cobra.Command, args []string) error {

		values := url.Values{}
		values.Set("grant_type", "client_credentials")

		if len(args) == 0 {
			log.Debug("Arguments have been passed, not using profile EPCC_CLIENT_ID and EPCC_CLIENT_SECRET")
			values.Set("client_id", config.Envs.EPCC_CLIENT_ID)
			values.Set("client_secret", config.Envs.EPCC_CLIENT_SECRET)
		}

		if len(args)%2 != 0 {
			return fmt.Errorf("invalid number of arguments supplied to login command, must be multiple of 2, not %v", len(args))
		}

		for i := 0; i < len(args); i += 2 {
			k := args[i]
			values.Set(k, args[i+1])
		}

		if authentication.IsCustomerTokenSet() {
			log.Infof("Destroying Customer Token as it should only be used with implicit tokens.")
			err := authentication.ClearCustomerToken()

			if err != nil {
				log.Warnf("Could not clear customer token: %v", err)
			}
		}

		token, err := authentication.GetAuthenticationToken(false, &values)

		if err != nil {
			return err
		}

		if token != nil {
			log.Infof("Successfully authenticated with client_credentials, session expires %s", time.Unix(token.Expires, 0).Format(time.RFC1123Z))
		} else {
			log.Warn("Did not successfully authenticate against the API")
		}

		return nil
	},
}

var loginImplicit = &cobra.Command{
	Use:   "implicit ([client_id <client_id>])",
	Short: "Login via implicit token",
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 || (args[0] != "client_id") {
			return completion.Complete(completion.Request{
				Type: completion.CompleteLoginClientID,
			})
		} else {
			return completion.Complete(completion.Request{
				Type: 0,
			})
		}
	},

	RunE: func(cmd *cobra.Command, args []string) error {

		values := url.Values{}
		values.Set("grant_type", "implicit")

		if len(args) == 0 {
			log.Debug("Arguments have been passed, not using profile EPCC_CLIENT_ID")
			values.Set("client_id", config.Envs.EPCC_CLIENT_ID)
		}

		if len(args)%2 != 0 {
			return fmt.Errorf("invalid number of arguments supplied to login command, must be multiple of 2, not %v", len(args))
		}

		for i := 0; i < len(args); i += 2 {
			k := args[i]
			values.Set(k, args[i+1])
		}

		token, err := authentication.GetAuthenticationToken(false, &values)

		if err != nil {
			return err
		}

		if token != nil {
			log.Infof("Successfully authenticated with implicit token, session expires %s", time.Unix(token.Expires, 0).Format(time.RFC1123Z))
		} else {
			log.Warn("Did not successfully authenticate against the API")
		}

		return nil
	},
}

var loginCustomer = &cobra.Command{
	Use:   "customer email <EMAIL> password <PASSWORD>",
	Short: "Obtain a customer token",
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {

		res, ok := resources.GetResourceByName("customer-token")

		if !ok {
			panic("Could not find customer token type")
		}

		if len(args)%2 == 0 {
			return completion.Complete(completion.Request{
				Type:     completion.CompleteAttributeKey,
				Verb:     completion.Create,
				Resource: res,
			})
		} else {
			usedAttributes := make(map[string]int)
			for i := 1; i < len(args); i = i + 2 {
				usedAttributes[args[i]] = 0
			}

			return completion.Complete(completion.Request{
				Type:       completion.CompleteAttributeValue,
				Verb:       completion.Create,
				Resource:   res,
				Attributes: usedAttributes,
			})
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		newArgs := make([]string, 0)
		newArgs = append(newArgs, "customer-token")
		newArgs = append(newArgs, args...)

		body, err := createInternal(newArgs)

		if err != nil {
			log.Warnf("Login not completed successfully")
			return err
		}

		apiToken := authentication.GetApiToken()

		if apiToken != nil {
			if apiToken.Identifier == "client_credentials" {
				log.Warnf("You are current logged in with client_credentials, please switch to implicit with `epcc login implicit` to use customer token correctly. Mixing client_credentials and customer token can lead to unintended results.")
			}
		}

		var customerTokenResponse *authentication.CustomerTokenResponse

		err = gojson.Unmarshal([]byte(body), &customerTokenResponse)

		if err != nil {
			return err
		}

		if customerTokenResponse != nil {

			// Get the customer so we have aliases where we need the id.
			getCustomerBody, err := getInternal([]string{"customer", customerTokenResponse.Data.CustomerId})

			if err != nil {
				log.Warnf("Could not retrieve customer")

			}

			if customerName, err := json.RunJQOnString(".data.name", getCustomerBody); customerName != nil && err == nil {
				if nameStr, ok := customerName.(string); ok {
					customerTokenResponse.AdditionalInfo.CustomerName = nameStr
				}
			}

			if customerEmail, err2 := json.RunJQOnString(".data.email", getCustomerBody); customerEmail != nil && err2 == nil {
				if emailStr, ok := customerEmail.(string); ok {
					customerTokenResponse.AdditionalInfo.CustomerEmail = emailStr
				}
			}

			log.Infof("Successfully authenticated as customer %s <%s>, session expires %s", customerTokenResponse.AdditionalInfo.CustomerName, customerTokenResponse.AdditionalInfo.CustomerEmail, time.Unix(customerTokenResponse.Data.Expires, 0).Format(time.RFC1123Z))

		} else {
			log.Warn("Did not successfully authenticate against the API")
		}

		authentication.SaveCustomerToken(*customerTokenResponse)

		return json.PrintJson(body)
	},
}
