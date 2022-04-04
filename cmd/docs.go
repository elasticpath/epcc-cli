package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/spf13/cobra"
	"os/exec"
	"runtime"
)

var docsCommand = &cobra.Command{
	Use:   "docs <resource>",
	Short: "Opens up API documentation for the resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			resource := resources.Resources[args[0]]
			url := resource.Docs

			switch runtime.GOOS {
			case "linux":
				exec.Command("xdg-open", url).Start()
			case "windows":
				exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
			case "darwin":
				exec.Command("open", url).Start()
			default:
				fmt.Errorf("unsupported platform")
			}
			return nil
			// return fmt.Errorf("This function is not implemented")

		} else {
			return fmt.Errorf("You must supply a resource type to the docs command")
		}
	},
}
