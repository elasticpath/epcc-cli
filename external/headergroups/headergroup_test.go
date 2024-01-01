package headergroups

import (
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAddHeaderGroup(t *testing.T) {
	// Fixture Setup
	ClearAllHeaderGroups()
	ClearAllHeaderAliasMappings()
	FlushHeaderGroups()

	// Execute SUT
	AddHeaderGroup("test", map[string]string{"foo": "bar"})

	// Verification
	require.Equal(t, []string{"test"}, GetAllHeaderGroups())
	require.Equal(t, map[string]string{"foo": "bar"}, GetAllHeaders())
}

func TestAddHeaderToHeaderGroupThatDoesNotExist(t *testing.T) {
	// Fixture Setup
	ClearAllHeaderGroups()
	ClearAllHeaderAliasMappings()
	FlushHeaderGroups()

	// Execute SUT
	AddHeaderToGroup("test", "foo", "bar")

	// Verification
	require.Equal(t, []string{"test"}, GetAllHeaderGroups())
	require.Equal(t, map[string]string{"foo": "bar"}, GetAllHeaders())
}

func TestAddHeaderGroupWithAlias(t *testing.T) {
	// Fixture Setup
	ClearAllHeaderGroups()
	ClearAllHeaderAliasMappings()
	FlushHeaderGroups()
	AddHeaderAliasMapping("FUZZY", "wuzzy")

	err := aliases.ClearAllAliases()

	require.NoError(t, err)

	aliases.SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "wuzzy"
	}
}`)

	// Execute SUT
	AddHeaderGroup("test", map[string]string{"Fuzzy": "id=123"})

	// Verification
	require.Equal(t, []string{"test"}, GetAllHeaderGroups())
	require.Equal(t, map[string]string{"Fuzzy": "123"}, GetAllHeaders())
}

func TestAddHeaderToHeaderGroupThatDoesExist(t *testing.T) {
	// Fixture Setup
	ClearAllHeaderGroups()
	ClearAllHeaderAliasMappings()
	FlushHeaderGroups()
	AddHeaderAliasMapping("FUZZY", "wuzzy")

	err := aliases.ClearAllAliases()

	require.NoError(t, err)

	aliases.SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "wuzzy"
	}
}`)

	AddHeaderGroup("test", map[string]string{"hello": "world"})

	// Execute SUT
	AddHeaderToGroup("test", "Fuzzy", "id=123")

	// Verification
	require.Equal(t, []string{"test"}, GetAllHeaderGroups())
	require.Equal(t, map[string]string{"Fuzzy": "123", "hello": "world"}, GetAllHeaders())
}

func TestAddHeaderToHeaderGroupWithAlias(t *testing.T) {
	// Fixture Setup
	ClearAllHeaderGroups()
	ClearAllHeaderAliasMappings()
	FlushHeaderGroups()
	AddHeaderGroup("test", map[string]string{"hello": "world"})

	// Execute SUT
	AddHeaderToGroup("test", "foo", "bar")

	// Verification
	require.Equal(t, []string{"test"}, GetAllHeaderGroups())
	require.Equal(t, map[string]string{"foo": "bar", "hello": "world"}, GetAllHeaders())
}

func TestRemoveHeaderGroupOnExistingGroup(t *testing.T) {
	// Fixture Setup
	ClearAllHeaderGroups()
	ClearAllHeaderAliasMappings()
	FlushHeaderGroups()
	AddHeaderGroup("test", map[string]string{"foo": "bar"})

	// Execute SUT
	RemoveHeaderGroup("test")

	// Verification
	require.Equal(t, []string{}, GetAllHeaderGroups())
	require.Equal(t, map[string]string{}, GetAllHeaders())
}

func TestRemoveHeaderGroupOnNonExistingGroup(t *testing.T) {
	// Fixture Setup
	ClearAllHeaderGroups()
	ClearAllHeaderAliasMappings()
	FlushHeaderGroups()
	AddHeaderGroup("test", map[string]string{"foo": "bar"})

	// Execute SUT
	RemoveHeaderGroup("does_not_exist")

	// Verification
	require.Equal(t, []string{"test"}, GetAllHeaderGroups())
	require.Equal(t, map[string]string{"foo": "bar"}, GetAllHeaders())
}

func TestRemoveHeaderFromGroupOnExistingGroup(t *testing.T) {
	// Fixture Setup
	ClearAllHeaderGroups()
	FlushHeaderGroups()
	AddHeaderGroup("test", map[string]string{"foo": "bar"})

	// Execute SUT
	RemoveHeaderFromGroup("test", "foo")

	// Verification
	require.Equal(t, []string{}, GetAllHeaderGroups())
	require.Equal(t, map[string]string{}, GetAllHeaders())
}

func TestRemoveHeaderFromGroupOnExistingGroupWithMultipleEntries(t *testing.T) {
	// Fixture Setup
	ClearAllHeaderGroups()
	ClearAllHeaderAliasMappings()
	FlushHeaderGroups()
	AddHeaderGroup("test", map[string]string{"foo": "bar", "hello": "world"})

	// Execute SUT
	RemoveHeaderFromGroup("test", "foo")

	// Verification
	require.Equal(t, []string{"test"}, GetAllHeaderGroups())
	require.Equal(t, map[string]string{"hello": "world"}, GetAllHeaders())
}

func TestRemoveHeaderFromGroupOnNonExisting(t *testing.T) {
	// Fixture Setup
	ClearAllHeaderGroups()
	ClearAllHeaderAliasMappings()
	FlushHeaderGroups()
	AddHeaderGroup("test", map[string]string{"foo": "bar"})

	// Execute SUT
	RemoveHeaderFromGroup("does_not_exist", "foo")

	// Verification
	require.Equal(t, []string{"test"}, GetAllHeaderGroups())
	require.Equal(t, map[string]string{"foo": "bar"}, GetAllHeaders())
}

func TestFlushHeaderGroup(t *testing.T) {
	// Fixture Setup
	ClearAllHeaderGroups()
	ClearAllHeaderAliasMappings()
	FlushHeaderGroups()
	AddHeaderGroup("test", map[string]string{"foo": "bar"})

	// Execute SUT
	FlushHeaderGroups()

	// Verification
	require.Equal(t, []string{"test"}, GetAllHeaderGroups())
	require.Equal(t, map[string]string{"foo": "bar"}, GetAllHeaders())
}
