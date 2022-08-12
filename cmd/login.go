package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/authentication"
	"github.com/elasticpath/epcc-cli/external/browser"
	"github.com/elasticpath/epcc-cli/external/completion"
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

var loginDocs = &cobra.Command{
	Use:       "docs",
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
	Use:   "client_credentials",
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
	Use:   "implicit",
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
