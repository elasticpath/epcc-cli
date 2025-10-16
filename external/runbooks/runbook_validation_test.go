package runbooks

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThatRunbookWithNoNameFails(t *testing.T) {

	// Fixture Setup

	// language=yaml
	runbookString := `
docs: http://localhost
`
	runbook, err := loadRunbookFromString(runbookString)
	assert.NoError(t, err, "Error should be nil")

	// Execute SUT
	err = ValidateRunbook(runbook)

	// Verification
	assert.ErrorContains(t, err, "Runbook has no name")
}

func TestThatRunbookWithNoActionsFails(t *testing.T) {

	// Fixture Setup

	// language=yaml
	runbookString := `
name: Foo
docs: http://localhost
`
	runbook, err := loadRunbookFromString(runbookString)
	assert.NoError(t, err, "Error should be nil")

	// Execute SUT
	err = ValidateRunbook(runbook)

	// Verification
	assert.ErrorContains(t, err, "number of actions is zero")
}

func TestThatRunbookWithEmptyActionFailsValidation(t *testing.T) {
	// Fixture Setup

	// language=yaml
	runbookString := `
name: unit-test-runbook
docs: "http://localhost"
actions:
  test-action:
    commands:
`

	runbook, err := loadRunbookFromString(runbookString)
	assert.NoError(t, err, "Error should be nil")

	// Execute SUT
	err = ValidateRunbook(runbook)

	// Verification
	assert.ErrorContains(t, err, "number of commands in action 'test-action' is zero")
}

func TestThatRunbookWithActionThatDoesNotStartWithEpccFails(t *testing.T) {
	// Fixture Setup

	// language=yaml
	runbookString := `
name: unit-test-runbook
docs: "http://localhost"
actions:
 test-action:
   commands:
   - "bar"
`

	runbook, err := loadRunbookFromString(runbookString)
	assert.NoError(t, err, "Error should be nil")

	// Execute SUT
	err = ValidateRunbook(runbook)

	// Verification
	assert.ErrorContains(t, err, "Each command needs be a recognized command")
}

func TestThatRunbookWithActionThatHasEpccButNoVerbFails(t *testing.T) {
	// Fixture Setup

	// language=yaml
	runbookString := `
name: unit-test-runbook
docs: "http://localhost"
actions:
 test-action:
   commands:
   - "epcc"
`

	runbook, err := loadRunbookFromString(runbookString)
	assert.NoError(t, err, "Error should be nil")

	// Execute SUT
	err = ValidateRunbook(runbook)

	// Verification
	assert.ErrorContains(t, err, "Each epcc command should be followed by a verb")
}

func TestThatRunbookCreateCommandPassesValidation(t *testing.T) {
	// Fixture Setup

	// language=yaml
	runbookString := `
name: unit-test-runbook
docs: "http://localhost"
actions:
 test-action:
   commands:
   - epcc create customer
`

	runbook, err := loadRunbookFromString(runbookString)
	assert.NoError(t, err, "Error should be nil")

	// Execute SUT
	err = ValidateRunbook(runbook)

	// Verification
	assert.NoError(t, err)
}

func TestThatRunbookGetCommandPassesValidation(t *testing.T) {
	// Fixture Setup

	// language=yaml
	runbookString := `
name: unit-test-runbook
docs: "http://localhost"
actions:
 test-action:
   commands:
   - epcc get customer
`

	runbook, err := loadRunbookFromString(runbookString)
	assert.NoError(t, err, "Error should be nil")

	// Execute SUT
	err = ValidateRunbook(runbook)

	// Verification
	assert.NoError(t, err)
}

func TestThatRunbookUpdateCommandPassesValidation(t *testing.T) {
	// Fixture Setup

	// language=yaml
	runbookString := `
name: unit-test-runbook
docs: "http://localhost"
actions:
 test-action:
   commands:
   - epcc update customer
`

	runbook, err := loadRunbookFromString(runbookString)
	assert.NoError(t, err, "Error should be nil")

	// Execute SUT
	err = ValidateRunbook(runbook)

	// Verification
	assert.NoError(t, err)
}

func TestThatRunbookDeleteCommandPassesValidation(t *testing.T) {
	// Fixture Setup

	// language=yaml
	runbookString := `
name: unit-test-runbook
docs: "http://localhost"
actions:
 test-action:
   commands:
   - epcc delete customer
`

	runbook, err := loadRunbookFromString(runbookString)
	assert.NoError(t, err, "Error should be nil")

	// Execute SUT
	err = ValidateRunbook(runbook)

	// Verification
	assert.NoError(t, err)
}

func TestThatRunbookDeleteAllCommandPassesValidation(t *testing.T) {
	// Fixture Setup

	// language=yaml
	runbookString := `
name: unit-test-runbook
docs: "http://localhost"
actions:
 test-action:
   commands:
   - epcc delete-all customer
`

	runbook, err := loadRunbookFromString(runbookString)
	assert.NoError(t, err, "Error should be nil")

	// Execute SUT
	err = ValidateRunbook(runbook)

	// Verification
	assert.NoError(t, err)
}

func TestThatRunbookSleepCommandPassesValidation(t *testing.T) {
	// Fixture Setup

	// language=yaml
	runbookString := `
name: unit-test-runbook
docs: "http://localhost"
actions:
 test-action:
   commands:
   - sleep 2
`

	runbook, err := loadRunbookFromString(runbookString)
	assert.NoError(t, err, "Error should be nil")

	// Execute SUT
	err = ValidateRunbook(runbook)

	// Verification
	assert.NoError(t, err)
}

func TestThatRunbookSleepCommandWithInvalidValueFailsValidation(t *testing.T) {
	// Fixture Setup

	// language=yaml
	runbookString := `
name: unit-test-runbook
docs: "http://localhost"
actions:
 test-action:
   commands:
   - sleep "a long time"
`

	runbook, err := loadRunbookFromString(runbookString)
	assert.NoError(t, err, "Error should be nil")

	// Execute SUT
	err = ValidateRunbook(runbook)

	// Verification
	assert.ErrorContains(t, err, "Invalid argument to sleep a long time")
}

func TestThatRunbookWithMultiLineActionOfAllDifferentCommandsPassesValidation(t *testing.T) {
	// Fixture Setup

	// language=yaml
	runbookString := `
name: unit-test-runbook
docs: "http://localhost"
actions:
 test-action:
   commands:
   - epcc create customer
   - epcc delete customer
   - epcc update customer
   - epcc delete-all customer
   - epcc get customer
   - sleep 2
`

	runbook, err := loadRunbookFromString(runbookString)
	assert.NoError(t, err, "Error should be nil")

	// Execute SUT
	err = ValidateRunbook(runbook)

	// Verification
	assert.NoError(t, err)
}

func TestThatRunbookWithUnknownMultiLineActionFailsValidation(t *testing.T) {
	// Fixture Setup

	// language=yaml
	runbookString := `
name: unit-test-runbook
docs: "http://localhost"
actions:
 test-action:
   commands:
   - epcc create customer
   - epcc delete customer
   - epcc update customer
   - epcc delete-all customer
   - foo
   - epcc get customer
   - sleep 2
   
`

	runbook, err := loadRunbookFromString(runbookString)
	assert.NoError(t, err, "Error should be nil")

	// Execute SUT
	err = ValidateRunbook(runbook)

	// Verification
	assert.ErrorContains(t, err, "Each command needs be a recognized command")
}

func TestThatRunbookWithPassesValidationWithVariableSubstitution(t *testing.T) {
	// Fixture Setup

	// language=yaml
	runbookString := `
name: unit-test-runbook
docs: "http://localhost"
actions:
  test-action:
    variables:
      action:
        type: STRING
        default: create
    commands:
    - epcc {{.action}} customer
`

	runbook, err := loadRunbookFromString(runbookString)
	assert.NoError(t, err, "Error should be nil")

	// Execute SUT
	err = ValidateRunbook(runbook)

	// Verification
	assert.NoError(t, err)
}

func TestThatRunbookWithFailsValidationWithVariableSubstitution(t *testing.T) {
	// Fixture Setup

	// language=yaml
	runbookString := `
name: unit-test-runbook
docs: "http://localhost"
actions:
  test-action:
    variables:
      action:
        type: STRING
        default: make-faster
    commands:
    - epcc {{.action}} customer
`

	runbook, err := loadRunbookFromString(runbookString)
	assert.NoError(t, err, "Error should be nil")

	// Execute SUT
	err = ValidateRunbook(runbook)

	// Verification
	assert.ErrorContains(t, err, "Each command needs to have a valid verb")
}

func TestThatRunbookWithInvalidTemplateFailsValidation(t *testing.T) {
	// Fixture Setup

	// language=yaml
	runbookString := `
name: unit-test-runbook
docs: "http://localhost"
actions:
 test-action:
   commands:
   - epcc customer {{ foo }}
`

	runbook, err := loadRunbookFromString(runbookString)
	assert.NoError(t, err, "Error should be nil")

	// Execute SUT
	err = ValidateRunbook(runbook)

	// Verification
	assert.ErrorContains(t, err, "error rendering template")
}

func TestThatRunbookWithValidTemplateThatFailsRenderingShouldReturnErrorValidation(t *testing.T) {
	// Fixture Setup

	// language=yaml
	runbookString := `
name: unit-test-runbook
docs: "http://localhost"
actions:
 test-action:
   commands:
   - epcc customer {{ fail "Sorry this isn't supported" }}
`

	runbook, err := loadRunbookFromString(runbookString)
	assert.NoError(t, err, "Error should be nil")

	// Execute SUT
	err = ValidateRunbook(runbook)

	// Verification
	assert.ErrorContains(t, err, "error rendering template")
}

func TestThatRunbookWithMismatchedQuotesFailsValidation(t *testing.T) {
	// Fixture Setup

	// language=yaml
	runbookString := `
name: unit-test-runbook
docs: "http://localhost"
actions:
 test-action:
   commands:
   - epcc create customer
   - sleep 1
   - |
     epcc create customer 
     epcc customer "Yo
`

	runbook, err := loadRunbookFromString(runbookString)
	assert.NoError(t, err, "Error should be nil")

	// Execute SUT
	err = ValidateRunbook(runbook)

	// Verification
	assert.ErrorContains(t, err, "expected closing quote")
}

func TestThatRunbookWithFailsValidationWithIntVariableNotHaveAnIntDefault(t *testing.T) {
	// Fixture Setup

	// language=yaml
	runbookString := `
name: unit-test-runbook
docs: "http://localhost"
actions:
  test-action:
    variables:
      time:
        type: INT
        default: make-faster
    commands:
    - sleep {{.time}}
`

	runbook, err := loadRunbookFromString(runbookString)
	assert.NoError(t, err, "Error should be nil")

	// Execute SUT
	err = ValidateRunbook(runbook)

	// Verification
	assert.ErrorContains(t, err, "error processing variable time, value make-faster is not an integer")
}

func TestThatRunbookWithFailsValidationWithUnknownVariableType(t *testing.T) {
	// Fixture Setup

	// language=yaml
	runbookString := `
name: unit-test-runbook
docs: "http://localhost"
actions:
  test-action:
    variables:
      time:
        type: magic
        default: make-faster
    commands:
    - sleep {{.time}}
`

	runbook, err := loadRunbookFromString(runbookString)
	assert.NoError(t, err, "Error should be nil")

	// Execute SUT
	err = ValidateRunbook(runbook)

	// Verification
	assert.ErrorContains(t, err, " error processing variable time, unknown type [magic] ")
}

func TestThatInitializeBuiltInRunbooksActuallyValidate(t *testing.T) {

	// Fixture Setup

	// language=yaml
	Reset()

	// Execute SUT
	InitializeBuiltInRunbooks()

	runbookNames := GetRunbookNames()

	for k, v := range runbooks {
		t.Run(k, func(t *testing.T) {
			assert.NoError(t, ValidateRunbook(&v))
		})
	}

	// Verification
	require.GreaterOrEqual(t, len(runbookNames), 1, "Expected that some runbooks should be loaded.")
}
