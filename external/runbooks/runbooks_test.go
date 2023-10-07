package runbooks

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestThatGetRunbookNamesReturnsNames(t *testing.T) {

	// Fixture Setup

	// language=yaml
	runbooks = map[string]Runbook{
		"test": {
			Name:           "foo",
			Description:    nil,
			Docs:           "https://www.google.ca",
			RunbookActions: nil,
		},
		"hello":   {},
		"goodbye": {},
	}
	// Execute SUT
	runbookNames := GetRunbookNames()

	// Verification
	require.Equal(t, []string{"goodbye", "hello", "test"}, runbookNames, "Expected that the names of runbooks should match and be sorted")

}

func TestThatGetRunbooksReturnsRunbooks(t *testing.T) {

	// Fixture Setup

	// language=yaml
	runbooks = map[string]Runbook{
		"test": {
			Name:           "foo",
			Description:    nil,
			Docs:           "https://www.google.ca",
			RunbookActions: nil,
		},
		"hello":   {},
		"goodbye": {},
	}
	// Execute SUT
	returnedRunbooks := GetRunbooks()

	// Verification
	require.Equal(t, runbooks, returnedRunbooks, "Expected that runbook objects should equal")
}

func TestThatAddRunbookWithNameAndValidYamlAddsRunbook(t *testing.T) {

	// Fixture Setup

	// language=yaml
	runbooks = map[string]Runbook{}

	validRunbook := `
name: test
docs: "http://localhost"
actions:
 test-action:
   commands:
   - epcc create customer
`

	// Execute SUT
	err := AddRunbookFromYaml(validRunbook)
	require.NoErrorf(t, err, "Should get no error when adding runbook")
	runbookNames := GetRunbookNames()

	// Verification
	require.Equal(t, []string{"test"}, runbookNames)
}

func TestThatInitializeBuiltInRunbooksActuallyLoadsRunbooks(t *testing.T) {

	// Fixture Setup

	// language=yaml
	Reset()

	// Execute SUT
	InitializeBuiltInRunbooks()

	runbookNames := GetRunbookNames()

	// Verification
	require.GreaterOrEqual(t, len(runbookNames), 1, "Expected that some runbooks should be loaded.")
}
