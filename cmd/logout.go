package cmd

import (
	"github.com/elasticpath/epcc-cli/external/authentication"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:          "logout",
	Short:        "Logout (Clears locally saved tokens)",
	SilenceUsage: false,
}

var logoutBearer = &cobra.Command{
	Use:     "api",
	Short:   "Logout of the API (Clears locally saved tokens _and_ prevents automatic login)",
	Aliases: []string{"bearer", "client_credentials", "implicit"},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := authentication.ClearApiToken()
		if err != nil {
			return err
		}

		err = authentication.DisableAutoLogin()
		if err != nil {
			return err
		}

		log.Info("Successfully logged out of the API, automatic login disabled")
		return nil
	},
}

var logoutCustomer = &cobra.Command{
	Use:   "customer",
	Short: "Destroys the customer token reverting to API only login",

	RunE: func(cmd *cobra.Command, args []string) error {

		if authentication.IsCustomerTokenSet() {
			err := authentication.ClearCustomerToken()
			if err != nil {
				return err
			}
			log.Info("Successfully destroyed the customer token")
			return nil
		} else {
			log.Info("No customer token found, you were already logged out.")
		}

		return nil
	},
}

var logoutAccountManagement = &cobra.Command{
	Use:   "account-management",
	Short: "Destroys the account management authentication token reverting to API only login",

	RunE: func(cmd *cobra.Command, args []string) error {

		if authentication.IsAccountManagementAuthenticationTokenSet() {
			err := authentication.ClearAccountManagementAuthenticationToken()
			if err != nil {
				return err
			}
			log.Info("Successfully destroyed the account management authentication token")
			return nil
		} else {
			log.Info("No account management authentication token found, you were already logged out.")
		}

		return nil
	},
}
