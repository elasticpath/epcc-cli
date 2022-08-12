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
	Short:   "Logout (Clears locally saved tokens _and_ prevents automatic login)",
	Aliases: []string{"bearer", "client_credentials", "implicit"},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := authentication.ClearApiToken()
		if err != nil {
			return err
		}

		authentication.DisableAutoLogin()
		if err != nil {
			return err
		}

		log.Info("Successfully logged out of the API, automatic login disabled")
		return nil
	},
}
