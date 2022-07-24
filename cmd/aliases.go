package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/resources"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"sort"
)

var aliasesCmd = &cobra.Command{
	Use:          "aliases",
	Short:        "Provides information about aliases that can be used",
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
				return fmt.Errorf("could not find resource information for resource: %s", args[0])
			}

			aliases := aliases.GetAliasesForJsonApiType(resource.JsonApiType)

			sortedAliasNames := make([]string, 0, len(aliases))

			for i := range aliases {
				sortedAliasNames = append(sortedAliasNames, i)
			}

			sort.Strings(sortedAliasNames)

			fmt.Printf("%40s || Values\n", "Alias Name")

			for _, alias := range sortedAliasNames {
				fmt.Printf("%40s => ID: %s", alias, aliases[alias].Id)
				if aliases[alias].Sku != "" {
					fmt.Printf(" Sku: %10s", aliases[alias].Sku)
				}

				if aliases[alias].Slug != "" {
					fmt.Printf(" Slug: %10s", aliases[alias].Slug)
				}

				fmt.Println()
			}

			return nil
		}
		return fmt.Errorf("you must supply a resource type to the aliases command")
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
		if err := aliases.ClearAllAliases(); err != nil {
			log.Info("Could not delete all resources")
			return err
		}
		log.Info("Successfully deleted all resources")
		return nil
	},
}
