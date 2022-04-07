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
			err = openDoc(args[0], "")
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
	resourceDoc, ok := resources.Resources[resource]
	if !ok {
		panic(fmt.Sprintf("Could not find resource %s", resource))
	}

	var err error
	switch verb {
	case "":
		if len(resourceDoc.Docs) < 1 {
			panic("You must supply a valid resource type to the docs command")
		}
		err = OpenUrl(resourceDoc.Docs)
	case "get-collection":
		if resourceDoc.GetCollectionInfo != nil && len(resourceDoc.GetCollectionInfo.Docs) < 1 {
			panic("couldn't find the document")
		}
		err = OpenUrl(resourceDoc.GetCollectionInfo.Docs)
	case "get":
		if resourceDoc.GetEntityInfo != nil && len(resourceDoc.GetEntityInfo.Docs) < 1 {
			panic("couldn't find the document")
		}
		err = OpenUrl(resourceDoc.GetEntityInfo.Docs)
	case "update":
		if resourceDoc.UpdateEntityInfo != nil && len(resourceDoc.UpdateEntityInfo.Docs) < 1 {
			panic("couldn't find the document")
		}
		err = OpenUrl(resourceDoc.UpdateEntityInfo.Docs)
	case "delete":
		if resourceDoc.DeleteEntityInfo != nil && len(resourceDoc.DeleteEntityInfo.Docs) < 1 {
			panic("couldn't find the document")
		}
		err = OpenUrl(resourceDoc.DeleteEntityInfo.Docs)
	case "create":
		if resourceDoc.CreateEntityInfo != nil && len(resourceDoc.CreateEntityInfo.Docs) < 1 {
			panic("couldn't find the document")
		}
		err = OpenUrl(resourceDoc.CreateEntityInfo.Docs)
	default:
		err = doDefault()

	}
	if err != nil {
		return err
	}
	return nil
}
func doDefault() error {
	return fmt.Errorf(" You must supply a resource type to the docs command")
}
