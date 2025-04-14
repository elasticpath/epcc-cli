package cmd

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/browser"
	"github.com/elasticpath/epcc-cli/external/completion"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/spf13/cobra"
)

var docsCommand = &cobra.Command{
	Use:   "docs <resource> [verb]",
	Short: "Opens up API documentation for the resource",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error = nil
		if len(args) != 0 {
			resource, ok := resources.GetResourceByName(args[0])
			if ok {
				switch len(args) {
				case 1:
					err = openDoc(resource, "")
				case 2:
					verb := args[1]
					err = openDoc(resource, verb)
				default:
					return fmt.Errorf("Unexpected number of arguments %d", len(args))
				}
			} else {
				if len(args) != 2 {
					return fmt.Errorf("you must supply a second argument because the first argument [%s] was not a resource", args[0])
				}

				resource, ok = resources.GetResourceByName(args[1])

				if !ok {
					return fmt.Errorf("neither argument was a resource [%v]", args)
				}

				return openDoc(resource, args[0])
			}

		} else {
			err = browser.OpenUrl("https://elasticpath.dev/docs/commerce-cloud/")
		}

		return err
	},

	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completion.Complete(completion.Request{
				Type: completion.CompletePluralResource + completion.CompleteCrudAction,
			})
		}

		if len(args) == 1 {
			_, ok := resources.GetResourceByName(args[0])
			if !ok {
				//first argument is not a resource, so the second must be
				return completion.Complete(completion.Request{
					Type: completion.CompletePluralResource,
				})
			} else {
				return completion.Complete(completion.Request{
					Type: completion.CompleteCrudAction,
				})
			}
		}
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	},
}

func openDoc(resourceDoc resources.Resource, verb string) error {
	var err error
	var info *resources.CrudEntityInfo

	docUrl := ""
	switch verb {
	case "":
		if len(resourceDoc.Docs) < 1 || resourceDoc.Docs == "n/a" {
			return fmt.Errorf("Could not open docs for resource '%s', no documentation available", resourceDoc.PluralName)
		}

		docUrl = resourceDoc.Docs
	case "get-collection":
		info = resourceDoc.GetCollectionInfo
	case "get":
		info = resourceDoc.GetEntityInfo
	case "update":
		info = resourceDoc.UpdateEntityInfo
	case "delete":
		info = resourceDoc.DeleteEntityInfo
	case "create":
		info = resourceDoc.CreateEntityInfo
	default:
		return fmt.Errorf("Unknown action for resource: [%s]", verb)
	}

	if info != nil {
		if len(info.Docs) < 1 || info.Docs == "n/a" {
			return fmt.Errorf("could not open docs for resource '%s', action'%s': no documentation available", resourceDoc.PluralName, verb)
		}
		docUrl = info.Docs
	}

	if docUrl == "" {
		return fmt.Errorf("no documentation available available for '%s', action '%s'", resourceDoc.PluralName, verb)
	}

	err = browser.OpenUrl(docUrl)

	if err != nil {
		return fmt.Errorf("error opening url: %w", err)
	}

	return nil
}
func doDefault() error {
	return fmt.Errorf("no documentation available")
}
