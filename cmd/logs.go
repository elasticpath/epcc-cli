package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/shared"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var LogsClear = &cobra.Command{
	Use:   "clear",
	Short: "Clears all HTTP request and response logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		os.RemoveAll(shared.LogDirectory)
		return nil
	},
}

var LogsList = &cobra.Command{
	Use:   "list",
	Short: "List all HTTP logs",
	RunE: func(cmd *cobra.Command, args []string) error {
		files := shared.AllFilesSortedByDate(shared.LogDirectory)
		for i := 0; i < len(files); i++ {
			name, _ := shared.Base64DecodeStripped(files[i].Name())
			fmt.Println(name)
		}
		return nil
	},
}

var LogsShow = &cobra.Command{
	Use:   "show <NUMBER>",
	Short: "Show HTTP logs for specific number",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		files := shared.AllFilesSortedByDate(shared.LogDirectory)
		for i := 0; i < len(files); i++ {
			name, _ := shared.Base64DecodeStripped(files[i].Name())
			segments := strings.Split(name, " ")
			if segments[0] == args[0] {
				content, err := os.ReadFile(shared.LogDirectory + "/" + files[i].Name())
				if err != nil {
					return err
				}
				fmt.Print(string(content))
				break
			}
		}

		return nil
	},
}

var Logs = &cobra.Command{Use: "logs"}
