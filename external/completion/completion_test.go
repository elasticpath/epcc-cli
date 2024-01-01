package completion

import (
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHeaderKeyWithNilValueCompletes(t *testing.T) {
	// Fixture Setup
	toComplete := "EP-"
	request := Request{
		Type:       CompleteHeaderKey,
		ToComplete: toComplete,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "EP-Beta-Features")
}

func TestHeaderKeyWithNonNilValueCompletes(t *testing.T) {
	// Fixture Setup
	toComplete := "X-Moltin"
	request := Request{
		Type:       CompleteHeaderKey,
		ToComplete: toComplete,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "X-Moltin-Currency")
}

func TestHeaderValueWithNilValueCompletesWithoutPanicing(t *testing.T) {
	// Fixture Setup
	toComplete := "ac"
	request := Request{
		Type:       CompleteHeaderValue,
		ToComplete: toComplete,
		Header:     "EP-Beta-Features",
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Empty(t, completions)

}

func TestHeaderValueWithNonNilValueCompletes(t *testing.T) {
	// Fixture Setup
	toComplete := "U"
	request := Request{
		Type:       CompleteHeaderValue,
		ToComplete: toComplete,
		Header:     "X-Moltin-Currency",
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "USD")
}
