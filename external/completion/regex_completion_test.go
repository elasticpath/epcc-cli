package completion

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSingleRegexReturnsCompletionOptionUpToFirstCaptureGroup(t *testing.T) {

	// Fixture Setup
	regex := "^custom_inputs\\.([a-zA-Z0-9-_]+)\\.name$"

	// Execute SUT
	rt := NewRegexCompletionTree()
	err := rt.AddRegex(regex)
	require.NoError(t, err)
	completionOptions, err := rt.GetCompletionOptions()

	// Verify
	require.NoError(t, err)
	require.Equal(t, []string{"custom_inputs."}, completionOptions)
}

func TestSingleRegexReturnsErrorIfMissingEndAnchor(t *testing.T) {

	// Fixture Setup
	regex := "^custom_inputs\\.([a-zA-Z0-9-_]+)\\.name"

	// Execute SUT
	rt := NewRegexCompletionTree()
	err := rt.AddRegex(regex)

	// Verify
	require.Error(t, err)
}

func TestTwoRegexReturnsCompletionOptionUpToFirstCaptureGroup(t *testing.T) {

	// Fixture Setup
	regexOne := "^custom_inputs\\.([a-zA-Z0-9-_]+)\\.name$"
	regexTwo := "^components\\.([a-zA-Z0-9-_]+)\\.min$"

	// Execute SUT
	rt := NewRegexCompletionTree()
	err := rt.AddRegex(regexOne)
	require.NoError(t, err)

	err = rt.AddRegex(regexTwo)
	require.NoError(t, err)

	completionOptions, err := rt.GetCompletionOptions()

	// Verify
	require.NoError(t, err)
	require.Len(t, completionOptions, 2)
	require.Contains(t, completionOptions, "custom_inputs.")
	require.Contains(t, completionOptions, "components.")
}

func TestAddExistingValueReturnsCompletionForFullValue(t *testing.T) {

	// Fixture Setup
	regex := "^components\\.([a-zA-Z0-9-_]+)\\.options\\[n]\\.id\\[n]$"
	rt := NewRegexCompletionTree()
	err := rt.AddRegex(regex)
	require.NoError(t, err)

	// Execute SUT
	err = rt.AddExistingValue("components.dogbed.options[0].id[0]")
	require.NoError(t, err)
	completionOptions, err := rt.GetCompletionOptions()

	// Verify
	require.NoError(t, err)
	require.Equal(t, []string{"components.dogbed.options[n].id[n]"}, completionOptions)
}

func TestAddExistingValueReturnsFullCompletionOptionsWithThreeOptions(t *testing.T) {

	// Fixture Setup

	rt := NewRegexCompletionTree()
	err := rt.AddRegex("^components\\.([a-zA-Z0-9-_]+)\\.options\\[n]\\.id\\[n]$")
	require.NoError(t, err)

	err = rt.AddRegex("^components\\.([a-zA-Z0-9-_]+)\\.options\\[n]\\.type$")
	require.NoError(t, err)

	err = rt.AddRegex("^components\\.([a-zA-Z0-9-_]+)\\.options\\[n]\\.quantity$")
	require.NoError(t, err)

	// Execute SUT
	err = rt.AddExistingValue("components.dogbed.options[0].id[0]")
	require.NoError(t, err)
	completionOptions, err := rt.GetCompletionOptions()

	// Verify
	require.NoError(t, err)
	require.Len(t, completionOptions, 3)
	require.Contains(t, completionOptions, "components.dogbed.options[n].id[n]")
	require.Contains(t, completionOptions, "components.dogbed.options[n].type")
	require.Contains(t, completionOptions, "components.dogbed.options[n].quantity")

}

func TestAddExistingValueReturnsFullCompletionOptionsWithThreeOptionsAndTwoExistingValues(t *testing.T) {

	// Fixture Setup

	rt := NewRegexCompletionTree()
	err := rt.AddRegex("^components\\.([a-zA-Z0-9-_]+)\\.options\\[n]\\.id\\[n]$")
	require.NoError(t, err)

	err = rt.AddRegex("^components\\.([a-zA-Z0-9-_]+)\\.options\\[n]\\.type$")
	require.NoError(t, err)

	err = rt.AddRegex("^components\\.([a-zA-Z0-9-_]+)\\.options\\[n]\\.quantity$")
	require.NoError(t, err)

	// Execute SUT
	err = rt.AddExistingValue("components.dogbed.options[0].id[0]")
	require.NoError(t, err)

	err = rt.AddExistingValue("components.dogbed.options[1].id[0]")
	require.NoError(t, err)

	err = rt.AddExistingValue("components.catbed.options[0].id[0]")
	require.NoError(t, err)

	completionOptions, err := rt.GetCompletionOptions()

	// Verify
	require.NoError(t, err)
	require.Len(t, completionOptions, 6)
	require.Contains(t, completionOptions, "components.dogbed.options[n].id[n]")
	require.Contains(t, completionOptions, "components.dogbed.options[n].type")
	require.Contains(t, completionOptions, "components.dogbed.options[n].quantity")
	require.Contains(t, completionOptions, "components.catbed.options[n].id[n]")
	require.Contains(t, completionOptions, "components.catbed.options[n].type")
	require.Contains(t, completionOptions, "components.catbed.options[n].quantity")

}

func TestAddExistingValueReturnsFullCompletionOptionsWithThreeOptionsAndTwoExistingValuesWithDistinctPrefixes(t *testing.T) {

	// Fixture Setup

	rt := NewRegexCompletionTree()
	err := rt.AddRegex("^components\\.([a-zA-Z0-9-_]+)\\.options\\[n]\\.id\\[n]$")
	require.NoError(t, err)

	err = rt.AddRegex("^components\\.([a-zA-Z0-9-_]+)\\.options\\[n]\\.type$")
	require.NoError(t, err)

	err = rt.AddRegex("^custom_inputs\\.([a-zA-Z0-9-_]+)\\.name$")
	require.NoError(t, err)

	// Execute SUT
	err = rt.AddExistingValue("components.dogbed.options[0].id[0]")
	require.NoError(t, err)

	err = rt.AddExistingValue("custom_inputs.foo.name")
	require.NoError(t, err)

	err = rt.AddExistingValue("components.catbed.options[0].id[0]")
	require.NoError(t, err)

	completionOptions, err := rt.GetCompletionOptions()

	// Verify
	require.NoError(t, err)
	require.Len(t, completionOptions, 5)
	require.Contains(t, completionOptions, "components.dogbed.options[n].id[n]")
	require.Contains(t, completionOptions, "components.dogbed.options[n].type")

	require.Contains(t, completionOptions, "components.catbed.options[n].id[n]")
	require.Contains(t, completionOptions, "components.catbed.options[n].type")
	require.Contains(t, completionOptions, "custom_inputs.foo.name")

}
