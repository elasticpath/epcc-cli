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

		resourceInfo := resources.GetPluralResources()
		sortedResourceNames := make([]string, 0, len(resourceInfo))

		for i := range resources.GetPluralResources() {
			sortedResourceNames = append(sortedResourceNames, i)
		}

		sort.Strings(sortedResourceNames)

		// Print resource list
		for _, resource := range sortedResourceNames {
			fmt.Printf("%s => json-type: %s\n", resource, resourceInfo[resource].JsonApiType)

			if resourceInfo[resource].GetCollectionInfo != nil {
				printCrudCommands(programName, "get", resourceInfo[resource].PluralName, resourceInfo[resource].GetCollectionInfo.Url)
			}

			if resourceInfo[resource].CreateEntityInfo != nil {
				printCrudCommands(programName, "create", resourceInfo[resource].SingularName, resourceInfo[resource].CreateEntityInfo.Url)
			}

			if resourceInfo[resource].GetEntityInfo != nil {
				printCrudCommands(programName, "get", resourceInfo[resource].SingularName, resourceInfo[resource].GetEntityInfo.Url)
			}

			if resourceInfo[resource].UpdateEntityInfo != nil {
				printCrudCommands(programName, "update", resourceInfo[resource].SingularName, resourceInfo[resource].UpdateEntityInfo.Url)
			}

			if resourceInfo[resource].DeleteEntityInfo != nil {
				printCrudCommands(programName, "delete", resourceInfo[resource].SingularName, resourceInfo[resource].DeleteEntityInfo.Url)
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
