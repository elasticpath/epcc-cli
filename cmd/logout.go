package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/globals"
	"github.com/spf13/cobra"
	"os"
)

var logout = &cobra.Command{
	Use:   "logout",
	Short: "Logout user",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("requires at least one arg")
		}
		apiArgName := args[0]
		// Can be extended for other user personas
		if apiArgName != API {
			return fmt.Errorf("argument is incorrect")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		var err error
		if _, err := os.Stat(globals.CredPath); err == nil {
			// Remove credentials after logout
			err = os.Remove(globals.CredPath)
		}
		if err != nil {
			return fmt.Errorf("User already logged out")
		}
		return err
	},
}
