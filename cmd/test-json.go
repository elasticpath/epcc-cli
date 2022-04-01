package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/spf13/cobra"
)

var testJson = &cobra.Command{
	Use:   "test-json [KEY_1] [VAL_1] [KEY_2] [VAL_2] ...",
	Short: "Prints the resulting json for what a command will look like",
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := json.ToJson(args)

		if res != "" {
			fmt.Println(res)
		}
		return err

	},
}
