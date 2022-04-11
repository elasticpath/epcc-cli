package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/shared"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var logsClear = &cobra.Command{
	Use:   "clear",
	Short: "Clears all Http logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		os.RemoveAll(shared.LogDirectory)
		return nil
	},
}

var logsList = &cobra.Command{
	Use:   "list",
	Short: "List All Http logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		files := shared.AllFilesSortedByDate(shared.LogDirectory)
		for i := 0; i < len(files); i++ {
			name, _ := shared.Base64DecodeStripped(files[i].Name())
			fmt.Println(name)
		}
		return nil
	},
}

var logsShow = &cobra.Command{
	Use:   "show <NUMBER>",
	Short: "Show Http logs for specific number",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		files := shared.AllFilesSortedByDate(shared.LogDirectory)
		for i := 0; i < len(files); i++ {
			name, _ := shared.Base64DecodeStripped(files[i].Name())
			segments := strings.Split(name, " ")
			if segments[0] == args[0] {
				fmt.Println(name)
				break
			}
		}

		return nil
	},
}

var logs = &cobra.Command{Use: "logs"}
