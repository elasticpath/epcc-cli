package cmd

import (
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"testing"
)

func init() {
	aliases.InitializeAliasDirectoryForTesting()
}

func TestGetCompletionForCollectionResourceReturnsStandardFields(t *testing.T) {

	// Fixture Setup
	rootCmd := &cobra.Command{}
	NewGetCommand(rootCmd)
	getCmd := getCommandForResource(rootCmd.Commands()[0], "accounts")

	require.NotNil(t, getCmd, "Get command for account should exist")

	// Execute SUT
	completionResult, _ := getCmd.ValidArgsFunction(getCmd, []string{}, "")

	// Verify
	require.Contains(t, completionResult, "sort")
	require.Contains(t, completionResult, "include")
	require.Contains(t, completionResult, "page[limit]")
	require.Contains(t, completionResult, "page[offset]")
	require.Contains(t, completionResult, "filter")
}

func TestGetCompletionForCollectionResourceReturnsValuesForStandardField(t *testing.T) {

	// Fixture Setup
	rootCmd := &cobra.Command{}
	NewGetCommand(rootCmd)
	getCmd := getCommandForResource(rootCmd.Commands()[0], "accounts")

	require.NotNil(t, getCmd, "Get command for account should exist")

	// Execute SUT
	completionResult, _ := getCmd.ValidArgsFunction(getCmd, []string{"sort"}, "")

	// Verify
	require.Contains(t, completionResult, "created_at")

}

func TestGetCompletionForCollectionWithParentResourceReturnsAlias(t *testing.T) {

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
	NewGetCommand(rootCmd)
	getCmd := getCommandForResource(rootCmd.Commands()[0], "account-addresses")

	require.NotNil(t, getCmd, "Get command for account should exist")

	// Execute SUT
	completionResult, _ := getCmd.ValidArgsFunction(getCmd, []string{}, "")

	// Verify
	require.Contains(t, completionResult, "name=John")

}

func TestGetCompletionForCollectionWithParentResourceReturnsStandardFields(t *testing.T) {

	// Fixture Setup
	rootCmd := &cobra.Command{}
	NewGetCommand(rootCmd)
	getCmd := getCommandForResource(rootCmd.Commands()[0], "account-addresses")

	require.NotNil(t, getCmd, "Get command for account should exist")

	// Execute SUT
	completionResult, _ := getCmd.ValidArgsFunction(getCmd, []string{"foo"}, "")

	// Verify
	require.Contains(t, completionResult, "sort")
	require.Contains(t, completionResult, "include")
	require.Contains(t, completionResult, "page[limit]")
	require.Contains(t, completionResult, "page[offset]")
	require.Contains(t, completionResult, "page[total_method]")
	require.Contains(t, completionResult, "filter")
}

func TestGetCompletionForCollectionResourceWithParentReturnsValuesForStandardField(t *testing.T) {

	// Fixture Setup
	rootCmd := &cobra.Command{}
	NewGetCommand(rootCmd)
	getCmd := getCommandForResource(rootCmd.Commands()[0], "account-addresses")

	require.NotNil(t, getCmd, "Get command for account should exist")

	// Execute SUT
	completionResult, _ := getCmd.ValidArgsFunction(getCmd, []string{"foo", "sort"}, "")

	// Verify
	require.Contains(t, completionResult, "created_at")

}

func TestGetArgFunctionForCollectionUrlWithNoParentsHasNoErrorWithNoArgs(t *testing.T) {
	// Fixture Setup
	resourceName := "accounts"

	rootCmd := &cobra.Command{}
	NewGetCommand(rootCmd)
	getCmd := getCommandForResource(rootCmd.Commands()[0], resourceName)

	// Execute SUT
	err := getCmd.Args(getCmd, []string{})

	// Verification
	require.NoError(t, err)
}

func TestGetArgFunctionForCollectionUrlWithParentHasErrorWithNoArgs(t *testing.T) {
	// Fixture Setup
	resourceName := "account-addresses"

	rootCmd := &cobra.Command{}
	NewGetCommand(rootCmd)
	getCmd := getCommandForResource(rootCmd.Commands()[0], resourceName)

	// Execute SUT
	err := getCmd.Args(getCmd, []string{})

	// Verification
	require.ErrorContains(t, err, "ACCOUNT_ID must be specified")
}

func TestGetArgFunctionForEntityUrlHasErrorWithNoArgs(t *testing.T) {
	// Fixture Setup
	resourceName := "account"

	rootCmd := &cobra.Command{}
	NewGetCommand(rootCmd)
	getCmd := getCommandForResource(rootCmd.Commands()[0], resourceName)

	// Execute SUT
	err := getCmd.Args(getCmd, []string{})

	// Verification
	require.ErrorContains(t, err, "ACCOUNT_ID must be specified")
}

func TestGetArgFunctionForCollectionUrlWithParentHasNoErrorWithArgs(t *testing.T) {
	// Fixture Setup
	resourceName := "account-addresses"

	rootCmd := &cobra.Command{}
	NewGetCommand(rootCmd)
	getCmd := getCommandForResource(rootCmd.Commands()[0], resourceName)

	// Execute SUT
	err := getCmd.Args(getCmd, []string{"foo"})

	// Verification
	require.NoError(t, err)
}

func TestGetArgFunctionForEntityUrlHasNoErrorWithArgs(t *testing.T) {
	// Fixture Setup
	resourceName := "account"

	rootCmd := &cobra.Command{}
	NewGetCommand(rootCmd)
	getCmd := getCommandForResource(rootCmd.Commands()[0], resourceName)

	// Execute SUT
	err := getCmd.Args(getCmd, []string{"foo"})

	// Verification
	require.NoError(t, err)
}
