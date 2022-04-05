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

type Request struct {
	Type     int
	Resource string
}

func Complete(c Request) ([]string, cobra.ShellCompDirective) {
	results := make([]string, 0)

	if c.Type&CompletePluralResource > 0 {
		for k := range resources.GetPluralResources() {
			results = append(results, k)
		}
	}

	if c.Type&CompleteSingularResource > 0 {
		for _, v := range resources.GetSingularResourceNames() {
			results = append(results, v)
		}
	}

	if c.Type&CompleteCrudAction > 0 {
		results = append(results, "create", "update", "delete", "get")
	}

	if c.Type&CompleteAttributeKey > 0 {
		// do something with :resources.Resources[c.Resource].Attributes
	}

	return results, cobra.ShellCompDirectiveNoFileComp
}
