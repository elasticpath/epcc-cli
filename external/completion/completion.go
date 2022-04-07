package completion

import (
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/spf13/cobra"
)

const (
	CompletePluralResource   = 1
	CompleteSingularResource = 2
	CompleteAttributeKey     = 4
	CompleteAttributeValue   = 8
	CompleteQueryParam       = 16
	CompleteCrudAction       = 32
)

const (
	Get    = 1
	Create = 2
	Update = 4
	Delete = 8
)

type Request struct {
	Type       int
	Resource   resources.Resource
	Attributes map[string]int
	Verb       int
}

func Complete(c Request) ([]string, cobra.ShellCompDirective) {
	results := make([]string, 0)

	if c.Type&CompletePluralResource > 0 {
		for k := range resources.GetPluralResources() {
			r, _ := resources.GetResourceByName(k) // Not worried about the bool here as resources come from the list already
			if c.Verb&Get > 0 {
				if r.GetCollectionInfo != nil {
					results = append(results, k)
				}
			} else {
				results = append(results, k)
			}
		}
	}

	if c.Type&CompleteSingularResource > 0 {
		for _, v := range resources.GetSingularResourceNames() {
			r, _ := resources.GetResourceByName(v) // Not worried about the bool here as resources come from the list already
			if c.Verb&Create > 0 {
				if r.CreateEntityInfo != nil {
					results = append(results, v)
				}
			} else if c.Verb&Update > 0 {
				if r.UpdateEntityInfo != nil {
					results = append(results, v)
				}
			} else if c.Verb&Delete > 0 {
				if r.DeleteEntityInfo != nil {
					results = append(results, v)
				}
			} else if c.Verb&Get > 0 {
				if r.GetEntityInfo != nil {
					results = append(results, v)
				}
			} else {
				results = append(results, v)
			}
		}
	}

	if c.Type&CompleteCrudAction > 0 {
		results = append(results, "create", "update", "delete", "get")
	}

	if c.Type&CompleteAttributeKey > 0 {
		for k := range c.Resource.Attributes {
			if _, ok := c.Attributes[k]; !ok {
				results = append(results, k)
			}
		}
	}

	return results, cobra.ShellCompDirectiveNoFileComp
}
