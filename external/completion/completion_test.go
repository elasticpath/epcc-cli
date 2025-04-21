package completion

import (
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"testing"
)

func init() {
	resources.PublicInit()
}

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

func TestAttributeValueWithNoTemplating(t *testing.T) {
	// Fixture Setup
	toComplete := ""
	acct := resources.MustGetResourceByName("password-profiles")
	request := Request{
		Type:       CompleteAttributeValue,
		Verb:       Create,
		ToComplete: toComplete,
		Attribute:  "username_format",
		Resource:   acct,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "any")
	require.Contains(t, completions, "email")
	require.Equal(t, 2, len(completions))
}

func TestAttributeValueWithTemplating(t *testing.T) {
	// Fixture Setup
	toComplete := ""
	acct := resources.MustGetResourceByName("password-profiles")
	request := Request{
		Type:           CompleteAttributeValue,
		Verb:           Create,
		ToComplete:     toComplete,
		Attribute:      "username_format",
		Resource:       acct,
		AllowTemplates: true,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "any")
	require.Contains(t, completions, "email")
	require.Contains(t, completions, `{{\ randAlphaNum\ |`)
	require.Contains(t, completions, `{{\ randAlphaNum\ }}`)
}

func TestAttributeValueWithTemplatingAndPipe(t *testing.T) {
	// Fixture Setup
	toComplete := "{{ randAlphaNum 3 | "
	acct := resources.MustGetResourceByName("password-profiles")
	request := Request{
		Type:           CompleteAttributeValue,
		Verb:           Create,
		ToComplete:     toComplete,
		Attribute:      "username_format",
		Resource:       acct,
		AllowTemplates: true,
	}

	// Exercise SUT
	completions, compDir := Complete(request)

	// Verify Results
	require.Equal(t, compDir, cobra.ShellCompDirectiveNoFileComp)
	require.Contains(t, completions, "any")
	require.Contains(t, completions, "email")
	require.Contains(t, completions, `{{\ randAlphaNum\ 3\ |\ upper\ |`)
	require.Contains(t, completions, `{{\ randAlphaNum\ 3\ |\ lower\ }}`)
}
