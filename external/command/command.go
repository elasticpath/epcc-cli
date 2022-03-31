package command

import "github.com/elasticpath/epcc-cli/config"

type Command struct {
	// The keyword the command should use
	Keyword string
	// A one-line description of the command
	Description string
	// The function that will be executed
	Execute func(cmds map[string]Command, cmd string, args []string, envs config.Env) int
}
