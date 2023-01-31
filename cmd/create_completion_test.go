package cmd

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateCompletionRetunsSomeTypes(t *testing.T) {

	// Fixture Setup

	// Execute SUT
	completionResult, _ := create.ValidArgsFunction(create, []string{}, "")

	// Verify
	require.Contains(t, completionResult, "customer")
	require.Contains(t, completionResult, "account")
}

func TestCreateCompletionReturnsSomeFields(t *testing.T) {

	// Fixture Setup

	// Execute SUT
	completionResult, _ := create.ValidArgsFunction(create, []string{"customer"}, "")

	// Verify
	require.Contains(t, completionResult, "name")
	require.Contains(t, completionResult, "email")
}

func TestCreateCompletionReturnsSomeFieldWhileExcludingUsedOnes(t *testing.T) {

	// Fixture Setup

	// Execute SUT
	completionResult, _ := create.ValidArgsFunction(create, []string{"customer", "name", "John"}, "")

	// Verify

	require.Contains(t, completionResult, "email")
	require.NotContains(t, completionResult, "name")
}
