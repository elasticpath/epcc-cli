package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/shared"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var logs = &cobra.Command{
	Use:   "logs <VERB> [NUMBER]",
	Short: "Returns Http logs",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		files := shared.AllFilesSortedByDate(shared.LogDirectory)
		if args[0] == "show" {
			if len(args) != 2 {
				return fmt.Errorf("Show command needs the number of log to show")
			}
		}
		if args[0] == "show" || args[0] == "list" {
			for i := 0; i < len(files); i++ {
				name, _ := shared.Base64DecodeStripped(files[i].Name())

				if args[0] == "show" {
					segments := strings.Split(name, " ")
					if segments[0] == args[1] {
						fmt.Println(name)
						break
					}
				} else if args[0] == "list" {
					fmt.Println(name)
				}
			}
		} else if args[0] == "clear" {
			os.RemoveAll(shared.LogDirectory)
		}

		return nil
	},
}
