package help

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/command"
	"sort"
)

var Command = command.Command{
	Keyword:     "help",
	Description: "Displays this screen",
	Execute: func(cmds map[string]command.Command, args []string) int {

		keys := make([]string, 0, len(cmds))
		for k := range cmds {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		fmt.Printf(`
Setup

The EPCC CLI tool uses environment variables for configuration and in particular a tool like https://direnv.net/ which
auto populates your shell with environment variables when you switch directories. This allows you to store a context in a folder,
and come back to it at any time.

Environment Variables

- EPCC_API_BASE_URL - The API endpoint that we will hit
- EPCC_CLIENT_ID - The client id (available in Commerce Manager)
- EPCC_CLIENT_SECRET - The client secret (available in Commerce Manager)
- EPCC_BETA_API_FEATURES - Beta features in the API we want to enable.

The following commands are supported:

`)

		for _, key := range keys {
			fmt.Printf("  %s - %s", cmds[key].Keyword, cmds[key].Description)
		}

		return 0
	},
}
