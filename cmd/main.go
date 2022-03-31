package main

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/command"
	commercemanager "github.com/elasticpath/epcc-cli/external/commerce-manager"
	"github.com/elasticpath/epcc-cli/external/help"
	_ "github.com/elasticpath/epcc-cli/external/resources"
	"os"
)

var commands = []command.Command{
	help.Command,
	commercemanager.Command,
}

func main() {
	envs := config.Env{}
	if err := env.Parse(&envs); err != nil {
		fmt.Printf("%+v\n", err)
	}

	argsWithoutProg := os.Args[1:]

	if (len(argsWithoutProg)) == 0 {
		fmt.Printf("No command specified")
		os.Exit(1)
	}

	commandToRun := argsWithoutProg[0]

	cmds := make(map[string]command.Command)

	for _, cmd := range commands {
		cmds[cmd.Keyword] = cmd
	}

	for _, cmd := range commands {
		if cmd.Keyword == commandToRun {
			argsWithoutCmd := argsWithoutProg[1:]
			os.Exit(cmd.Execute(cmds, argsWithoutProg[0], argsWithoutCmd, envs))
		}
	}

	fmt.Printf("Unknown command %s specified", commandToRun)
	os.Exit(0)

}
