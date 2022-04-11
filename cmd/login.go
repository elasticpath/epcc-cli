package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/authentication"
	"github.com/elasticpath/epcc-cli/globals"
	"github.com/spf13/cobra"
	"os"
)

const (
	API          = "api"
	ClientId     = "client_id"
	ClientSecret = "client_secret"
)

var login = &cobra.Command{
	Use:   "login",
	Short: "Authenticate by providing credentials.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("requires at least one arg")
		}
		apiArgName := args[0]
		// Can be extended for other user personas
		if apiArgName != API {
			return fmt.Errorf("argument is incorrect")
		}

		if len(args) < 2 {
			return fmt.Errorf("requires client_id argument")
		}

		clientIdArgName := args[1]
		if clientIdArgName != ClientId {
			return fmt.Errorf("argument is incorrect")
		}

		if len(args) > 3 {
			clientSecretArgName := args[3]
			if clientSecretArgName != ClientSecret {
				return fmt.Errorf("argument is incorrect")
			}
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		globals.EpccClientId = args[2]
		if len(args) > 3 {
			globals.EpccClientSecret = args[4]
		}
		globals.NewLogin = true
		_, err := authentication.GetAuthenticationToken()

		// Persist credentials to a file after successful login
		if err == nil {
			s1 := globals.EpccClientId
			s2 := globals.EpccClientSecret
			s := []byte(s1 + ";" + s2)
			err = os.WriteFile(globals.CredPath, s, 0644)
		}

		return err
	},
}
