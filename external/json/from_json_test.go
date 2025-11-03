package json

import (
	"testing"

	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/stretchr/testify/require"
)

func init() {
	aliases.InitializeAliasDirectoryForTesting()
	resources.PublicInit()
}

func TestFromJsonWithEmptyString(t *testing.T) {
	// Fixture Setup
	// language=json
	json := `{}`

	// Execute SUT
	result, err := FromJson(json)

	// Verification
	require.NoError(t, err)
	require.Empty(t, result)
}

func TestFromJsonWithSimpleObject(t *testing.T) {
	// Fixture Setup
	//language=json
	json := `{
     "data": {
        "type": "account",
        "name": "Ron"
    }
  }`

	// Execute SUT
	result, err := FromJson(json)

	// Verification
	require.NoError(t, err)
	require.Equal(t, result, []string{"name", "\"Ron\"", "type", "\"account\""})
}

func TestFromJsonWithSimpleObjectWithAttributes(t *testing.T) {
	// Fixture Setup
	//language=json
	json := `{
     "data": {
        "type": "account",
        "attributes": {
           "name": "Ron"
        }
    }
  }`

	// Execute SUT
	result, err := FromJson(json)

	// Verification
	require.NoError(t, err)
	require.Equal(t, result, []string{"name", `"Ron"`, "type", `"account"`})
}

func TestFromJsonWithSimpleObjectWithArrayAttributes(t *testing.T) {
	// Fixture Setup
	//language=json
	json := `{
     "data": {
        "type": "account",
        "attributes": {
           "names": ["Ron", "Ulysses", "Swanson"]
        }
    }
  }`

	// Execute SUT
	result, err := FromJson(json)

	// Verification
	require.NoError(t, err)
	require.Equal(t, result, []string{"names[0]", `"Ron"`, "names[1]", `"Ulysses"`, "names[2]", `"Swanson"`, "type", `"account"`})
}
