package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/spf13/cobra"
)

var docsCommand = &cobra.Command{
	Use:   "docs [RESOURCE] [ID_1]",
	Short: "Opens up API documentation for the resource",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		arg := len(args)

		switch arg {
		case 0:
			return fmt.Errorf("You must supply a resource type to the docs command")
		case 1:
			resource := resources.Resources[args[0]]
			if len(resource.Docs) > 0 {
				openDoc(args[0], "")
			} else {
				return fmt.Errorf("You must supply a valid resource type to the docs command")
			}
		case 2:
			resource := args[0]
			verb := args[1]
			openDoc(resource, verb)
		default:
			return fmt.Errorf("unsupported platform")
		}
		return nil
	},
}

func openDoc(resource string, verb string) error {
	resourceDoc := resources.Resources[resource]

	switch verb {
	case "":
		OpenUrl(resourceDoc.Docs)
	case "get-collection":
		OpenUrl(resourceDoc.GetCollectionInfo.Docs)
	case "get-entity":
		OpenUrl(resourceDoc.GetEntityInfo.Docs)
	case "update-entity":
		OpenUrl(resourceDoc.UpdateEntityInfo.Docs)
	case "delete-entity":
		OpenUrl(resourceDoc.DeleteEntityInfo.Docs)
	case "create-entity":
		OpenUrl(resourceDoc.CreateEntityInfo.Docs)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return nil
}
