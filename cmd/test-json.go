package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var testJson = &cobra.Command{
	Use:   "test-json [KEY_1] [VAL_1] [KEY_2] [VAL_2] ...",
	Short: "Prints the resulting json for what a command will look like",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("This function is not implemented")
	},
}
