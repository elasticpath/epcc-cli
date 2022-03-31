package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var docsCommand = &cobra.Command{
	Use:   "docs <resource>",
	Short: "Opens up API documentation for the resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("This function is not implemented")
	},
}
