package cmd

import (
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"testing"
)

func init() {
	resources.PublicInit()
}

func TestCreateCompletionReturnsSomeFields(t *testing.T) {

	// Fixture Setup
	rootCmd := &cobra.Command{}
	NewCreateCommand(rootCmd)
	create := getCommandForResource(rootCmd.Commands()[0], "account")

	require.NotNil(t, create, "Create command for account should exist")

	// Execute SUT
	completionResult, _ := create.ValidArgsFunction(create, []string{}, "")

	// Verify
	require.Contains(t, completionResult, "name")
	require.Contains(t, completionResult, "legal_name")
}

func TestCreateCompletionReturnsSomeFieldWhileExcludingUsedOnes(t *testing.T) {

	// Fixture Setup
	rootCmd := &cobra.Command{}
	NewCreateCommand(rootCmd)
	create := getCommandForResource(rootCmd.Commands()[0], "account")

	require.NotNil(t, create, "Create command for account should exist")

	// Execute SUT
	completionResult, _ := create.ValidArgsFunction(create, []string{"name", "John"}, "")

	// Verify
	require.Contains(t, completionResult, "legal_name")
	require.Contains(t, completionResult, "registration_id")
	require.NotContains(t, completionResult, "name")
}

func TestCreateCompletionReturnsFirstElementParentId(t *testing.T) {

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
	NewCreateCommand(rootCmd)
	createCmd := getCommandForResource(rootCmd.Commands()[0], "account-address")

	require.NotNil(t, createCmd, "Update command for account-addresses should exist")

	// Execute SUT
	completionResult, _ := createCmd.ValidArgsFunction(createCmd, []string{}, "")

	// Verify
	require.Contains(t, completionResult, "name=John")
}

func TestCreateCompletionReturnsAValidAttributeKey(t *testing.T) {

	// Fixture Setup
	rootCmd := &cobra.Command{}
	NewCreateCommand(rootCmd)
	createCmd := getCommandForResource(rootCmd.Commands()[0], "account-address")

	require.NotNil(t, createCmd, "Update command for account-addresses should exist")

	// Execute SUT
	completionResult, _ := createCmd.ValidArgsFunction(createCmd, []string{"name=John"}, "")

	// Verify
	require.Contains(t, completionResult, "county")
	require.Contains(t, completionResult, "city")
}

func TestCreateCompletionReturnsAValidAttributeKeyThatHasNotBeenUsed(t *testing.T) {

	// Fixture Setup
	rootCmd := &cobra.Command{}
	NewCreateCommand(rootCmd)
	createCmd := getCommandForResource(rootCmd.Commands()[0], "account-address")

	require.NotNil(t, createCmd, "Update command for account-addresses should exist")

	// Execute SUT
	completionResult, _ := createCmd.ValidArgsFunction(createCmd, []string{"name=John", "city", "Whitewood"}, "")

	// Verify
	require.Contains(t, completionResult, "county")
	require.NotContains(t, completionResult, "city")
}

func TestCreateCompletionReturnsAValidAttributeValue(t *testing.T) {
	// Fixture Setup
	rootCmd := &cobra.Command{}
	NewCreateCommand(rootCmd)
	createCmd := getCommandForResource(rootCmd.Commands()[0], "integration")

	require.NotNil(t, createCmd, "Update command for integrations should exist")

	// Execute SUT
	completionResult, _ := createCmd.ValidArgsFunction(createCmd, []string{"integration_type"}, "")

	// Verify
	require.Contains(t, completionResult, "webhook")
	require.Contains(t, completionResult, "aws_sqs")
}

func TestCreateArgFunctionForEntityUrlHasNoErrorWithNoArgs(t *testing.T) {
	// Fixture Setup
	resourceName := "account"

	rootCmd := &cobra.Command{}
	NewCreateCommand(rootCmd)
	createCmd := getCommandForResource(rootCmd.Commands()[0], resourceName)

	// Execute SUT
	err := createCmd.Args(createCmd, []string{})

	// Verification
	require.NoError(t, err)
}

func TestCreateArgFunctionForEntityUrlWithParentIdHasErrorWithNoArgs(t *testing.T) {
	// Fixture Setup
	resourceName := "account-address"

	rootCmd := &cobra.Command{}
	NewCreateCommand(rootCmd)
	createCmd := getCommandForResource(rootCmd.Commands()[0], resourceName)

	// Execute SUT
	err := createCmd.Args(createCmd, []string{})
	// Verification
	require.ErrorContains(t, err, "ACCOUNT_ID must be specified")
}
