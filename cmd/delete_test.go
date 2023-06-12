package cmd

import (
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDeleteCompletionReturnsFirstElementParentId(t *testing.T) {
	// Fixture Setup
	err := aliases.ClearAllAliases()

	require.NoError(t, err)

	aliases.SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "account",
		"name": "John"
	}
}`)

	rootCmd := &cobra.Command{}
	NewDeleteCommand(rootCmd)
	deleteCmd := getCommandForResource(rootCmd.Commands()[0], "account-address")

	require.NotNil(t, deleteCmd, "Delete command for account-addresses should exist")

	// Execute SUT
	completionResult, _ := deleteCmd.ValidArgsFunction(deleteCmd, []string{}, "")

	// Verify
	require.Contains(t, completionResult, "name=John")
}

func TestDeleteCompletionReturnsSecondElementId(t *testing.T) {
	// Fixture Setup
	err := aliases.ClearAllAliases()

	require.NoError(t, err)

	aliases.SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "address"
	}
}`)

	rootCmd := &cobra.Command{}
	NewDeleteCommand(rootCmd)
	deleteCmd := getCommandForResource(rootCmd.Commands()[0], "account-address")

	require.NotNil(t, deleteCmd, "Delete command for account-addresses should exist")

	// Execute SUT
	completionResult, _ := deleteCmd.ValidArgsFunction(deleteCmd, []string{"name=John"}, "")

	// Verify
	require.Contains(t, completionResult, "id=123")
}

func TestDeleteCompletionReturnsAnValidAttributeKey(t *testing.T) {

	// Fixture Setup
	rootCmd := &cobra.Command{}
	NewDeleteCommand(rootCmd)
	deleteCmd := getCommandForResource(rootCmd.Commands()[0], "account-address")

	require.NotNil(t, deleteCmd, "Delete command for account-addresses should exist")

	// Execute SUT
	completionResult, _ := deleteCmd.ValidArgsFunction(deleteCmd, []string{"name=John", "id=123"}, "")

	// Verify
	require.Contains(t, completionResult, "county")
	require.Contains(t, completionResult, "city")
}

func TestDeleteCompletionReturnsAnValidAttributeKeyThatHasNotBeenUsed(t *testing.T) {

	// Fixture Setup
	rootCmd := &cobra.Command{}
	NewDeleteCommand(rootCmd)
	deleteCmd := getCommandForResource(rootCmd.Commands()[0], "account-address")

	require.NotNil(t, deleteCmd, "Delete command for account-addresses should exist")

	// Execute SUT
	completionResult, _ := deleteCmd.ValidArgsFunction(deleteCmd, []string{"name=John", "id=123", "city", "Lumsden"}, "")

	// Verify
	require.Contains(t, completionResult, "county")
	require.NotContains(t, completionResult, "city")
}

func TestDeleteCompletionReturnsAnValidAttributeValue(t *testing.T) {

	// Fixture Setup
	rootCmd := &cobra.Command{}
	NewDeleteCommand(rootCmd)
	deleteCmd := getCommandForResource(rootCmd.Commands()[0], "integration")

	require.NotNil(t, deleteCmd, "Delete command for account-addresses should exist")

	// Execute SUT
	completionResult, _ := deleteCmd.ValidArgsFunction(deleteCmd, []string{"id=123", "integration_type"}, "")

	// Verify
	require.Contains(t, completionResult, "webhook")
	require.Contains(t, completionResult, "aws_sqs")
}

func TestDeleteArgFunctionForEntityUrlHasErrorWithNoArgs(t *testing.T) {
	// Fixture Setup
	resourceName := "account"

	rootCmd := &cobra.Command{}
	NewDeleteCommand(rootCmd)
	deleteCmd := getCommandForResource(rootCmd.Commands()[0], resourceName)

	// Execute SUT
	err := deleteCmd.Args(deleteCmd, []string{})

	// Verification
	require.ErrorContains(t, err, "ACCOUNT_ID must be specified")
}

func TestDeleteArgFunctionForEntityUrlWithParentIdHasErrorWithNoArgs(t *testing.T) {
	// Fixture Setup
	resourceName := "account-address"

	rootCmd := &cobra.Command{}
	NewDeleteCommand(rootCmd)
	deleteCmd := getCommandForResource(rootCmd.Commands()[0], resourceName)

	// Execute SUT
	err := deleteCmd.Args(deleteCmd, []string{})

	// Verification
	require.ErrorContains(t, err, "ACCOUNT_ID, ACCOUNT_ADDRESS_ID must be specified")

}

func TestDeleteArgFunctionForEntityUrlHasNoErrorWithArgs(t *testing.T) {
	// Fixture Setup
	resourceName := "account"

	rootCmd := &cobra.Command{}
	NewDeleteCommand(rootCmd)
	deleteCmd := getCommandForResource(rootCmd.Commands()[0], resourceName)

	// Execute SUT
	err := deleteCmd.Args(deleteCmd, []string{"foo"})

	// Verification
	require.NoError(t, err)

}

func TestDeleteArgFunctionForEntityUrlWithParentIdHasErrorWithOneArgOnly(t *testing.T) {
	// Fixture Setup
	resourceName := "account-address"

	rootCmd := &cobra.Command{}
	NewDeleteCommand(rootCmd)
	deleteCmd := getCommandForResource(rootCmd.Commands()[0], resourceName)

	// Execute SUT
	err := deleteCmd.Args(deleteCmd, []string{"foo"})

	// Verification
	require.ErrorContains(t, err, "ACCOUNT_ADDRESS_ID must be specified")
	require.NotContains(t, err.Error(), "ACCOUNT_ID must be specified")
}

func TestDeleteArgFunctionForEntityUrlWithParentIdHasNoErrorWithArgs(t *testing.T) {
	// Fixture Setup
	resourceName := "account-address"

	rootCmd := &cobra.Command{}
	NewDeleteCommand(rootCmd)
	deleteCmd := getCommandForResource(rootCmd.Commands()[0], resourceName)

	// Execute SUT
	err := deleteCmd.Args(deleteCmd, []string{"foo", "bar"})

	// Verification
	require.NoError(t, err)
}
