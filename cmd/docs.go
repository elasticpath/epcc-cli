package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/elasticpath/epcc-cli/shared"
	"github.com/spf13/cobra"
)

var docsCommand = &cobra.Command{
	Use:   "docs <resource>",
	Short: "Opens up API documentation for the resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		if len(args) != 0 {
			resource, ok := resources.GetResourceByName(args[0])
			if !ok {
				return fmt.Errorf("Could not find resource information for resource: %s", args[0])
			}
			switch len(args) {
			case 1:
				err = openDoc(resource, "")
			case 2:
				verb := args[1]
				err = openDoc(resource, verb)
			default:
				return doDefault()
			}
		}
		if err != nil {
			return fmt.Errorf(err.Error())
		}
		return nil
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

func openDoc(resourceDoc resources.Resource, verb string) error {
	var err error
	switch verb {
	case "":
		if len(resourceDoc.Docs) < 1 {
			err = doDefault()
		}
		err = shared.OpenUrl(resourceDoc.Docs)
	case "get-collection":
		if resourceDoc.GetCollectionInfo != nil && len(resourceDoc.GetCollectionInfo.Docs) < 1 {
			err = doDefault()
		}
		err = shared.OpenUrl(resourceDoc.GetCollectionInfo.Docs)
	case "get":
		if resourceDoc.GetEntityInfo != nil && len(resourceDoc.GetEntityInfo.Docs) < 1 {
			err = doDefault()
		}
		err = shared.OpenUrl(resourceDoc.GetEntityInfo.Docs)
	case "update":
		if resourceDoc.UpdateEntityInfo != nil && len(resourceDoc.UpdateEntityInfo.Docs) < 1 {
			err = doDefault()
		}
		err = shared.OpenUrl(resourceDoc.UpdateEntityInfo.Docs)
	case "delete":
		if resourceDoc.DeleteEntityInfo != nil && len(resourceDoc.DeleteEntityInfo.Docs) < 1 {
			err = doDefault()
		}
		err = shared.OpenUrl(resourceDoc.DeleteEntityInfo.Docs)
	case "create":
		if resourceDoc.CreateEntityInfo != nil && len(resourceDoc.CreateEntityInfo.Docs) < 1 {
			err = doDefault()
		}
		err = shared.OpenUrl(resourceDoc.CreateEntityInfo.Docs)
	default:
		return fmt.Errorf("Could not find verb %s", verb)

	}
	if err != nil {
		return err
	}
	return nil
}
func doDefault() error {
	return fmt.Errorf(" You must supply a resource type to the docs command")
}
