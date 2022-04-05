package completion

import (
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/spf13/cobra"
)

const (
	CompleteResource       = 0
	CompleteAttributeKey   = 1
	CompleteAttributeValue = 2
	CompleteQueryParam     = 4
	CompleteCrudAction     = 8
)

type CompletionRequest struct {
	Type     int
	Resource string
}

func Complete(c CompletionRequest) ([]string, cobra.ShellCompDirective) {
	results := make([]string, 0)

	if c.Type&CompleteResource > 0 {
		for k, _ := range resources.Resources {
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
