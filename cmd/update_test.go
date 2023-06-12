package cmd

import (
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUpdateCompletionReturnsFirstElementParentId(t *testing.T) {

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
	NewUpdateCommand(rootCmd)
	updateCmd := getCommandForResource(rootCmd.Commands()[0], "account-address")

	require.NotNil(t, updateCmd, "Update command for account-addresses should exist")

	// Execute SUT
	completionResult, _ := updateCmd.ValidArgsFunction(updateCmd, []string{}, "")

	// Verify
	require.Contains(t, completionResult, "name=John")
}

func TestUpdateCompletionReturnsSecondElementId(t *testing.T) {

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
	NewUpdateCommand(rootCmd)
	updateCmd := getCommandForResource(rootCmd.Commands()[0], "account-address")

	require.NotNil(t, updateCmd, "Update command for account-addresses should exist")

	// Execute SUT
	completionResult, _ := updateCmd.ValidArgsFunction(updateCmd, []string{"name=John"}, "")

	// Verify
	require.Contains(t, completionResult, "id=123")
}

func TestUpdateCompletionReturnsAnValidAttributeKey(t *testing.T) {

	// Fixture Setup
	rootCmd := &cobra.Command{}
	NewUpdateCommand(rootCmd)
	updateCmd := getCommandForResource(rootCmd.Commands()[0], "account-address")

	require.NotNil(t, updateCmd, "Update command for account-addresses should exist")

	// Execute SUT
	completionResult, _ := updateCmd.ValidArgsFunction(updateCmd, []string{"name=John", "id=123"}, "")

	// Verify
	require.Contains(t, completionResult, "county")
	require.Contains(t, completionResult, "city")
}

func TestUpdateCompletionReturnsAnValidAttributeKeyThatHasNotBeenUsed(t *testing.T) {

	// Fixture Setup
	rootCmd := &cobra.Command{}
	NewUpdateCommand(rootCmd)
	updateCmd := getCommandForResource(rootCmd.Commands()[0], "account-address")

	require.NotNil(t, updateCmd, "Update command for account-addresses should exist")

	// Execute SUT
	completionResult, _ := updateCmd.ValidArgsFunction(updateCmd, []string{"name=John", "id=123", "city", "Aylesbury"}, "")

	// Verify
	require.Contains(t, completionResult, "county")
	require.NotContains(t, completionResult, "city")
}

func TestUpdateCompletionReturnsAnValidAttributeValue(t *testing.T) {

	// Fixture Setup
	rootCmd := &cobra.Command{}
	NewUpdateCommand(rootCmd)
	updateCmd := getCommandForResource(rootCmd.Commands()[0], "authentication-realm")

	require.NotNil(t, updateCmd, "Update command for account-addresses should exist")

	// Execute SUT
	completionResult, _ := updateCmd.ValidArgsFunction(updateCmd, []string{"id=123", "duplicate_email_policy"}, "")

	// Verify
	require.Contains(t, completionResult, "allowed")
	require.Contains(t, completionResult, "api_only")

}

func TestUpdateArgFunctionForEntityUrlHasErrorWithNoArgs(t *testing.T) {
	// Fixture Setup
	resourceName := "account"

	rootCmd := &cobra.Command{}
	NewUpdateCommand(rootCmd)
	updateCmd := getCommandForResource(rootCmd.Commands()[0], resourceName)

	// Execute SUT
	err := updateCmd.Args(updateCmd, []string{})

	// Verification
	require.ErrorContains(t, err, "ACCOUNT_ID must be specified")
}

func TestUpdateArgFunctionForEntityUrlWithParentIdHasErrorWithNoArgs(t *testing.T) {
	// Fixture Setup
	resourceName := "account-address"

	rootCmd := &cobra.Command{}
	NewUpdateCommand(rootCmd)
	updateCmd := getCommandForResource(rootCmd.Commands()[0], resourceName)

	// Execute SUT
	err := updateCmd.Args(updateCmd, []string{})

	// Verification
	require.ErrorContains(t, err, "ACCOUNT_ID, ACCOUNT_ADDRESS_ID must be specified")

}

func TestUpdateArgFunctionForEntityUrlHasNoErrorWithArgs(t *testing.T) {
	// Fixture Setup
	resourceName := "account"

	rootCmd := &cobra.Command{}
	NewUpdateCommand(rootCmd)
	updateCmd := getCommandForResource(rootCmd.Commands()[0], resourceName)

	// Execute SUT
	err := updateCmd.Args(updateCmd, []string{"foo"})

	// Verification
	require.NoError(t, err)

}

func TestUpdateArgFunctionForEntityUrlWithParentIdHasErrorWithOneArgOnly(t *testing.T) {
	// Fixture Setup
	resourceName := "account-address"

	rootCmd := &cobra.Command{}
	NewUpdateCommand(rootCmd)
	updateCmd := getCommandForResource(rootCmd.Commands()[0], resourceName)

	// Execute SUT
	err := updateCmd.Args(updateCmd, []string{"foo"})

	// Verification
	require.ErrorContains(t, err, "ACCOUNT_ADDRESS_ID must be specified")
	require.NotContains(t, err.Error(), "ACCOUNT_ID must be specified")
}

func TestUpdateArgFunctionForEntityUrlWithParentIdHasNoErrorWithArgs(t *testing.T) {
	// Fixture Setup
	resourceName := "account-address"

	rootCmd := &cobra.Command{}
	NewUpdateCommand(rootCmd)
	updateCmd := getCommandForResource(rootCmd.Commands()[0], resourceName)

	// Execute SUT
	err := updateCmd.Args(updateCmd, []string{"foo", "bar"})

	// Verification
	require.NoError(t, err)
}
