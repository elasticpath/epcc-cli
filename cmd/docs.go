package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/spf13/cobra"
)

var docsCommand = &cobra.Command{
	Use:   "docs <resource>",
	Short: "Opens up API documentation for the resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			resource := resources.Resources[args[0]]
			if len(resource.Docs) > 0 {
				url := resource.Docs
				err := OpenUrl(url)
				if err != nil {
					return nil
				}
			} else {
				return fmt.Errorf("You must supply a valid resource type to the docs command")
			}
		}
		return fmt.Errorf("You must supply a resource type to the docs command")
	},
}
