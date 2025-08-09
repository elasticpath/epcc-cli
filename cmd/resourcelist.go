package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/config"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"sort"

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

		hiddenResources := map[string]struct{}{}
		e := config.GetEnv()
		for _, v := range e.EPCC_CLI_DISABLE_RESOURCES {
			hiddenResources[v] = struct{}{}
		}

		for i := range resources.GetPluralResources() {
			sortedResourceNames = append(sortedResourceNames, i)
		}

		sort.Strings(sortedResourceNames)

		// Print resource list
		for _, resource := range sortedResourceNames {

			r := resourceInfo[resource]
			if _, ok := hiddenResources[r.SingularName]; ok {
				log.Tracef("Hiding resource %s", r.SingularName)
				continue
			}

			if _, ok := hiddenResources[r.PluralName]; ok {
				log.Tracef("Hiding resource %s", r.PluralName)
				continue
			}

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

func printCrudCommands(programName, operation, resource, url string) error {
	// Determine number of template variables required
	idCount, err := resources.GetNumberOfVariablesNeeded(url)

	if err != nil {
		return err
	}

	// Generate resource operation with ID requirement string
	resourceOperationString := operation + " " + resource

	for idNum := 1; idNum <= idCount; idNum++ {
		resourceOperationString += fmt.Sprintf(" [ID%d]", idNum)
	}

	fmt.Printf("%16s %-70s ==> %s\n", programName, resourceOperationString, url)

	return nil
}
