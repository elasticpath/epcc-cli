package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/spf13/cobra"
)

var docsCommand = &cobra.Command{
	Use:   "docs [RESOURCE] [VERB]",
	Short: "Opens up API documentation for the resource",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		switch len(args) {
		case 0:
			return fmt.Errorf("You must supply a resource type to the docs command")
		case 1:
			resource := resources.Resources[args[0]]
			if len(resource.Docs) > 0 {
				err = openDoc(args[0], "")
			} else {
				return fmt.Errorf("You must supply a valid resource type to the docs command")
			}
		case 2:
			resource := args[0]
			verb := args[1]
			err = openDoc(resource, verb)

		default:
			return doDefault()
		}
		if err != nil {
			return doDefault()
		}
		return nil
	},
}

func openDoc(resource string, verb string) error {
	resourceDoc := resources.Resources[resource]
	var err error
	switch verb {
	case "":
		err = OpenUrl(resourceDoc.Docs)
	case "get-collection":
		if len(resourceDoc.GetCollectionInfo.Docs) < 1 {
			return doDefault()
		}
		err = OpenUrl(resourceDoc.GetCollectionInfo.Docs)
	case "get-entity":
		if len(resourceDoc.GetEntityInfo.Docs) < 1 {
			return doDefault()
		}
		err = OpenUrl(resourceDoc.GetEntityInfo.Docs)
	case "update-entity":
		if len(resourceDoc.UpdateEntityInfo.Docs) < 1 {
			return doDefault()
		}
		err = OpenUrl(resourceDoc.UpdateEntityInfo.Docs)
	case "delete-entity":
		if len(resourceDoc.DeleteEntityInfo.Docs) < 1 {
			return doDefault()
		}
		err = OpenUrl(resourceDoc.DeleteEntityInfo.Docs)
	case "create-entity":
		if len(resourceDoc.CreateEntityInfo.Docs) < 1 {
			return doDefault()
		}
		err = OpenUrl(resourceDoc.CreateEntityInfo.Docs)
	default:
		doDefault()
	}
	if err != nil {
		return err
	}
	return nil
}
func doDefault() error {
	return fmt.Errorf("unsupported platform")
}
