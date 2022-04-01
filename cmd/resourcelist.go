package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/spf13/cobra"
)

var resourceListCommand = &cobra.Command{
	Use:   "resource-list",
	Short: "Lists all resources and supported operations for each",
	Args:  cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine program name
		programName := filepath.Base(os.Args[0])

		// Generate sorted list of resource names
		sortedResourceNames := make([]string, 0, len(resources.Resources))

		for i := range resources.Resources {
			sortedResourceNames = append(sortedResourceNames, i)
		}

		sort.Strings(sortedResourceNames)

		// Print resource list
		for _, resource := range sortedResourceNames {
			fmt.Printf("%s => json-type: %s\n", resource, resources.Resources[resource].JsonApiType)

			if resources.Resources[resource].GetCollectionInfo != nil {
				printCrudCommands(programName, "get", resource, resources.Resources[resource].GetCollectionInfo.Url)
			}

			if resources.Resources[resource].CreateEntityInfo != nil {
				printCrudCommands(programName, "create", resource, resources.Resources[resource].CreateEntityInfo.Url)
			}

			if resources.Resources[resource].GetEntityInfo != nil {
				printCrudCommands(programName, "get", resource, resources.Resources[resource].GetEntityInfo.Url)
			}

			if resources.Resources[resource].UpdateEntityInfo != nil {
				printCrudCommands(programName, "update", resource, resources.Resources[resource].UpdateEntityInfo.Url)
			}

			if resources.Resources[resource].DeleteEntityInfo != nil {
				printCrudCommands(programName, "delete", resource, resources.Resources[resource].DeleteEntityInfo.Url)
			}

			fmt.Printf("\n")
		}

		return nil
	},
}

func printCrudCommands(programName, operation, resource, url string) {
	numIdRequired := strings.Count(url, "%")
	fmt.Printf("%16s %-70s ==> %s\n", programName, operation+" "+resource+strings.Repeat(" [ID]", numIdRequired), url)
}
