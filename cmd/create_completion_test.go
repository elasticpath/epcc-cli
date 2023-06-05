package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestCreateCompletionReturnsSomeFields(t *testing.T) {

	// Fixture Setup
	rootCmd := &cobra.Command{}
	NewCreateCommand(rootCmd)
	create := getCommandForResource(rootCmd.Commands()[0], "customer")

	require.NotNil(t, create, "Create command for customer should exist")

	// Execute SUT
	completionResult, _ := create.ValidArgsFunction(create, []string{}, "")

	// Verify
	require.Contains(t, completionResult, "name")
	require.Contains(t, completionResult, "email")
}

func TestCreateCompletionReturnsSomeFieldWhileExcludingUsedOnes(t *testing.T) {

	// Fixture Setup
	rootCmd := &cobra.Command{}
	NewCreateCommand(rootCmd)
	create := getCommandForResource(rootCmd.Commands()[0], "customer")

	require.NotNil(t, create, "Create command for customer should exist")

	// Execute SUT
	completionResult, _ := create.ValidArgsFunction(create, []string{"name", "John"}, "")

	// Verify
	require.Contains(t, completionResult, "email")
	require.NotContains(t, completionResult, "name")
}

func getCommandForResource(cmd *cobra.Command, res string) *cobra.Command {
	for _, c := range cmd.Commands() {
		if strings.HasPrefix(c.Use, res+" ") {
			return c
		}
	}
	return nil
}
