package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/spf13/cobra"
)

var docsCommand = &cobra.Command{
	Use:   "docs <resource>",
	Short: "Opens up API documentation for the resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			resource, ok := resources.GetResourceByName(args[0])
			if !ok {
				return fmt.Errorf("Could not find resource information for resource: %s", args[0])
			}
			if len(resource.Docs) > 0 {
				url := resource.Docs
				err := OpenUrl(url)
				return err
			} else {
				return fmt.Errorf("You must supply a valid resource type to the docs command")
			}
		}
		return fmt.Errorf("You must supply a resource type to the docs command")
	},

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completion.Complete(completion.Request{
				Type: completion.CompleteResource,
			})
		}

		return []string{}, cobra.ShellCompDirectiveNoFileComp
	},
}
