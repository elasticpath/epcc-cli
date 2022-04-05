package completion

import (
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/spf13/cobra"
)

const (
	CompleteResource       = 1
	CompleteAttributeKey   = 2
	CompleteAttributeValue = 4
	CompleteQueryParam     = 8
	CompleteCrudAction     = 16
)

type Request struct {
	Type     int
	Resource string
}

func Complete(c Request) ([]string, cobra.ShellCompDirective) {
	results := make([]string, 0)

	if c.Type&CompleteResource > 0 {
		for k := range resources.Resources {
			results = append(results, k)
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