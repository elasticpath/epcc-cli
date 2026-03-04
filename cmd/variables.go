package cmd

import (
	"fmt"
	"sort"

	"github.com/elasticpath/epcc-cli/external/variables"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var variablesCmd = &cobra.Command{
	Use:          "variables",
	Short:        "Manage variables extracted from API responses",
	SilenceUsage: false,
}

var variableListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all stored variables",
	RunE: func(cmd *cobra.Command, args []string) error {
		allVars := variables.GetAllVariables()

		if len(allVars) == 0 {
			fmt.Println("No variables set.")
			return nil
		}

		names := make([]string, 0, len(allVars))
		for name := range allVars {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			fmt.Printf("%s = %s\n", name, allVars[name])
		}

		return nil
	},
}

var variableClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all stored variables",
	RunE: func(cmd *cobra.Command, args []string) error {
		variables.ClearAllVariables()
		variables.FlushVariables()
		log.Info("Successfully cleared all variables")
		return nil
	},
}

var variableSetCmd = &cobra.Command{
	Use:   "set <name> <value>",
	Short: "Set a variable to a value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		variables.SetVariable(args[0], args[1])
		log.Infof("Set variable %q = %q", args[0], args[1])
		return nil
	},
}

// NewVariablesCommand adds the variables command tree to a parent command.
// Used by runbook command pool to enable `epcc variables set` in scripts.
func NewVariablesCommand(parent *cobra.Command) {
	vCmd := &cobra.Command{
		Use:          "variables",
		Short:        "Manage variables extracted from API responses",
		SilenceUsage: false,
	}
	vCmd.AddCommand(
		&cobra.Command{
			Use:   "list",
			Short: "Lists all stored variables",
			RunE:  variableListCmd.RunE,
		},
		&cobra.Command{
			Use:   "clear",
			Short: "Clear all stored variables",
			RunE:  variableClearCmd.RunE,
		},
		&cobra.Command{
			Use:   "set <name> <value>",
			Short: "Set a variable to a value",
			Args:  cobra.ExactArgs(2),
			RunE:  variableSetCmd.RunE,
		},
	)
	parent.AddCommand(vCmd)
}
