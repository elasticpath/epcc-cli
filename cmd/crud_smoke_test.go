package cmd

import (
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCrudOnAResource(t *testing.T) {

	httpclient.Initialize(1, 60)

	cmd := getTestCommand()
	cmd.SetArgs([]string{"create", "account", "name", "Test", "legal_name", "Test", "--output-jq", ".data.id", "--save-as-alias", "my_test_alias"})
	err := cmd.Execute()
	require.NoError(t, err)

	foo := aliases.GetAliasesForJsonApiTypeAndAlternates("account", []string{})
	_, ok := foo["my_test_alias"]

	require.True(t, ok, "Expected that my_test_alias exists in the set of aliases :(")

	cmd = getTestCommand()
	cmd.SetArgs([]string{"get", "account", "name=Test", "--output-jq", ".data.name"})
	err = cmd.Execute()
	require.NoError(t, err)

	cmd = getTestCommand()
	cmd.SetArgs([]string{"get", "accounts", "--output-jq", ".data[].name"})
	err = cmd.Execute()
	require.NoError(t, err)

	cmd = getTestCommand()
	cmd.SetArgs([]string{"update", "account", "name=Test", "legal_name", "Test Update", "--output-jq", ".data.legal_name"})
	err = cmd.Execute()
	require.NoError(t, err)

	cmd = getTestCommand()
	cmd.SetArgs([]string{"delete", "account", "name=Test"})
	err = cmd.Execute()
	require.NoError(t, err)

	// Error because this UUID doesn't exist
	cmd = getTestCommand()
	cmd.SetArgs([]string{"delete", "account", "6e7e2cdb-ff61-45a9-956b-c9dfc28d11d0"})
	err = cmd.Execute()
	require.Error(t, err)

	// No error because of argument
	cmd = getTestCommand()
	cmd.SetArgs([]string{"delete", "account", "6e7e2cdb-ff61-45a9-956b-c9dfc28d11d0", "--allow-404"})
	err = cmd.Execute()
	require.NoError(t, err)

	// Missing required arg
	cmd = getTestCommand()
	cmd.SetArgs([]string{"create", "account"})
	err = cmd.Execute()
	require.Error(t, err)

	// Resource doesn't exist
	cmd = getTestCommand()
	cmd.SetArgs([]string{"update", "account", "6e7e2cdb-ff61-45a9-956b-c9dfc28d11d0"})
	err = cmd.Execute()
	require.Error(t, err)

	// Resource doesn't exist
	cmd = getTestCommand()
	cmd.SetArgs([]string{"delete", "account", "6e7e2cdb-ff61-45a9-956b-c9dfc28d11d0"})
	err = cmd.Execute()
	require.Error(t, err)

}

func getTestCommand() *cobra.Command {
	testRootCmd := &cobra.Command{
		SilenceUsage: true,
	}

	NewCreateCommand(testRootCmd)
	NewGetCommand(testRootCmd)
	NewUpdateCommand(testRootCmd)
	NewDeleteCommand(testRootCmd)

	initConfig()

	return testRootCmd

}
