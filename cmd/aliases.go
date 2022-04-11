package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/profiles"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/spf13/cobra"
	"os"
	"sort"
)

var aliasesCmd = &cobra.Command{
	Use:          "aliases",
	SilenceUsage: false,
}

var aliasListCmd = &cobra.Command{
	Use:   "list <resource>",
	Short: "Lists all aliases for a resource",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			resource, ok := resources.GetResourceByName(args[0])
			if !ok {
				return fmt.Errorf("Could not find resource information for resource: %s", args[0])
			}

			aliases := aliases.GetAliasesForJsonApiType(resource.JsonApiType)

			sortedAliasNames := make([]string, 0, len(aliases))

			for i := range aliases {
				sortedAliasNames = append(sortedAliasNames, i)
			}

			sort.Strings(sortedAliasNames)

			for _, alias := range sortedAliasNames {
				fmt.Printf("%40s => %s\n", alias, aliases[alias])
			}

			return nil
		}
		return fmt.Errorf("You must supply a resource type to the aliases command")
	},

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completion.Complete(completion.Request{
				Type: completion.CompletePluralResource,
			})
		}

		return []string{}, cobra.ShellCompDirectiveNoFileComp
	},
}

var aliasClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "clear all aliases",
	RunE: func(cmd *cobra.Command, args []string) error {
		profileDirectory := profiles.GetProfileDataBaseURL()
		os.RemoveAll(profileDirectory)
		return nil
	},
}
