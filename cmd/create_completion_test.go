package cmd

import (
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateCompletionReturnsSomeTypes(t *testing.T) {

	// Fixture Setup
	rootCmd := &cobra.Command{}
	NewCreateCommand(rootCmd)
	create := rootCmd.Commands()[0]

	// Execute SUT
	completionResult, _ := create.ValidArgsFunction(create, []string{}, "")

	// Verify
	require.Contains(t, completionResult, "customer")
	require.Contains(t, completionResult, "account")
}

func TestCreateCompletionReturnsSomeFields(t *testing.T) {

	// Fixture Setup
	rootCmd := &cobra.Command{}
	NewCreateCommand(rootCmd)
	create := rootCmd.Commands()[0]

	// Execute SUT
	completionResult, _ := create.ValidArgsFunction(create, []string{"customer"}, "")

	// Verify
	require.Contains(t, completionResult, "name")
	require.Contains(t, completionResult, "email")
}

func TestCreateCompletionReturnsSomeFieldWhileExcludingUsedOnes(t *testing.T) {

	// Fixture Setup
	rootCmd := &cobra.Command{}
	NewCreateCommand(rootCmd)
	create := rootCmd.Commands()[0]

	// Execute SUT
	completionResult, _ := create.ValidArgsFunction(create, []string{"customer", "name", "John"}, "")

	// Verify

	require.Contains(t, completionResult, "email")
	require.NotContains(t, completionResult, "name")
}
