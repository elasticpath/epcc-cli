package aliases

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func init() {
	InitializeAliasDirectoryForTesting()
}

func TestSavedAliasIsReturnedInAllAliasesForSingleResponse(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo"
	}
}`)

	aliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})

	// Verification
	require.Len(t, aliases, 2, "There should be %d typeToAliasNameToIdMap in map not %d", 2, len(aliases))

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "123", aliases["id=123"].Id)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "123", aliases["last_read=entity"].Id)
}

func TestSavedAliasAppendsAndPreservesPreviousUnrelatedAliases(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo"
	}
}`)
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "456",
		"type": "foo"
	}
}`)

	aliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})

	// Verification

	require.Len(t, aliases, 3, "There should be %d typeToAliasNameToIdMap in map not %d", 3, len(aliases))

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "123", aliases["id=123"].Id)

	require.Contains(t, aliases, "id=456")
	require.Equal(t, "456", aliases["id=456"].Id)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "456", aliases["last_read=entity"].Id)
}

func TestSavedAliasIsReplacedWhenNewEntityHasTheSameAttributeValue(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo",
		"name": "Alpha"
	}
}`)
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "456",
		"type": "foo",
		"name":"Alpha"
	}
}`)

	aliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})

	// Verification

	require.Len(t, aliases, 4, "There should be %d typeToAliasNameToIdMap in map not %d", 4, len(aliases))

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "123", aliases["id=123"].Id)

	require.Contains(t, aliases, "id=456")
	require.Equal(t, "456", aliases["id=456"].Id)

	require.Contains(t, aliases, "name=Alpha")
	require.Equal(t, "456", aliases["name=Alpha"].Id)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "456", aliases["last_read=entity"].Id)
}

func TestSavedAliasIsReplacedWhenSameEntityHasANewValue(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo",
		"name": "Alpha"
	}
}`)
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo",
		"name":"Beta"
	}
}`)

	aliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})

	// Verification

	require.Len(t, aliases, 3, "There should be %d typeToAliasNameToIdMap in map not %d", 3, len(aliases))

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "123", aliases["id=123"].Id)

	require.Contains(t, aliases, "name=Beta")
	require.Equal(t, "123", aliases["name=Beta"].Id)
	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "123", aliases["last_read=entity"].Id)
}

func TestThatLastReadAliasesAreNotReplacedWhenSeenInADifferentContext(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo",
		"name": "Alpha"
	}
}`)
	SaveAliasesForResources(
		// language=JSON
		`
{	
	"data": [{
		"id": "123",
		"type": "foo",
		"name":"Beta"
		}]
}`)

	aliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})

	// Verification

	require.Len(t, aliases, 4, "There should be %d typeToAliasNameToIdMap in map not %d", 3, len(aliases))

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "123", aliases["id=123"].Id)

	require.Contains(t, aliases, "name=Beta")
	require.Equal(t, "123", aliases["name=Beta"].Id)
	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "123", aliases["last_read=entity"].Id)
	require.Contains(t, aliases, "last_read=array[0]")
	require.Equal(t, "123", aliases["last_read=array[0]"].Id)
}

func TestDeleteAliasByIdDeletesAnAlias(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo",
		"name": "Steve",
		"sku": "456",
		"slug": "foo-123"
	}
}`)
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "456",
		"type": "foo",
		"name": "Steve",
		"sku": "456",
		"slug": "foo-456"
	}
}`)

	// Execute SUT

	DeleteAliasesById("123", "foo")

	aliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})

	// Verification

	require.Len(t, aliases, 5, "There should be %d typeToAliasNameToIdMap in map not %d", 5, len(aliases))

	require.Contains(t, aliases, "id=456")
	require.Equal(t, "456", aliases["id=456"].Id)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "456", aliases["last_read=entity"].Id)

	require.Contains(t, aliases, "name=Steve")
	require.Equal(t, "456", aliases["name=Steve"].Id)

	require.Contains(t, aliases, "sku=456")
	require.Equal(t, "456", aliases["sku=456"].Id)

	require.Contains(t, aliases, "slug=foo-456")
	require.Equal(t, "456", aliases["slug=foo-456"].Id)
}

func TestDeleteAliasByIdDeletesAnAliasOnly(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo"
	}
}`)
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "456",
		"type": "foo"
	}
}`)

	// Execute SUT

	DeleteAliasesById("123", "foo")

	aliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})

	// Verification

	require.Len(t, aliases, 2, "There should be %d typeToAliasNameToIdMap in map not %d", 2, len(aliases))

	require.Contains(t, aliases, "id=456")
	require.Equal(t, "456", aliases["id=456"].Id)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "456", aliases["last_read=entity"].Id)
}

func TestAllAliasesAreReturnedInAllAliasesForArrayResponse(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": [{
		"id": "123",
		"type": "foo"
	}, {
		"id": "456",
		"type": "foo"
	}
	]
}
`)

	aliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})

	// Verification

	require.Len(t, aliases, 4, "There should be %d typeToAliasNameToIdMap in map not %d", 4, len(aliases))

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "123", aliases["id=123"].Id)

	require.Contains(t, aliases, "id=456")
	require.Equal(t, "456", aliases["id=456"].Id)

	require.Contains(t, aliases, "last_read=array[0]")
	require.Equal(t, "123", aliases["last_read=array[0]"].Id)

	require.Contains(t, aliases, "last_read=array[1]")
	require.Equal(t, "456", aliases["last_read=array[1]"].Id)
}

func TestSavedAliasIsReturnedForAnEmailInLegacyObjectResponse(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"email": "test@test.com",
		"type": "foo"
	}
}`)

	aliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})

	// Verification

	require.Len(t, aliases, 3, "There should be %d typeToAliasNameToIdMap in map not %d", 3, len(aliases))

	require.Contains(t, aliases, "email=test@test.com")
	require.Equal(t, "123", aliases["email=test@test.com"].Id)

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "123", aliases["id=123"].Id)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "123", aliases["last_read=entity"].Id)
}

func TestSavedAliasIsReturnedForAnSkuInLegacyObjectResponse(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"sku": "test",
		"type": "foo"
	}
}`)

	aliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})

	// Verification

	require.Len(t, aliases, 3, "There should be %d typeToAliasNameToIdMap in map not %d", 3, len(aliases))

	require.Contains(t, aliases, "sku=test")
	require.Equal(t, "123", aliases["sku=test"].Id)

	require.Contains(t, aliases, "sku=test")
	require.Equal(t, "test", aliases["sku=test"].Sku)

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "123", aliases["id=123"].Id)

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "test", aliases["id=123"].Sku)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "123", aliases["last_read=entity"].Id)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "test", aliases["last_read=entity"].Sku)
}

func TestSavedAliasIsReturnedForAnCodeInLegacyObjectResponse(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"code": "hello",
		"type": "foo"
	}
}`)

	aliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})

	// Verification

	require.Len(t, aliases, 3, "There should be %d typeToAliasNameToIdMap in map not %d", 3, len(aliases))

	require.Contains(t, aliases, "code=hello")
	require.Equal(t, "123", aliases["code=hello"].Id)

	require.Contains(t, aliases, "code=hello")
	require.Equal(t, "hello", aliases["code=hello"].Code)

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "123", aliases["id=123"].Id)

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "hello", aliases["id=123"].Code)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "123", aliases["last_read=entity"].Id)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "hello", aliases["last_read=entity"].Code)
}

func TestSavedAliasIsReturnedForASlugInLegacyObjectResponse(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"slug": "test",
		"type": "foo"
	}
}`)

	aliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})

	// Verification

	require.Len(t, aliases, 3, "There should be %d typeToAliasNameToIdMap in map not %d", 3, len(aliases))

	require.Contains(t, aliases, "slug=test")
	require.Equal(t, "123", aliases["slug=test"].Id)

	require.Contains(t, aliases, "slug=test")
	require.Equal(t, "test", aliases["slug=test"].Slug)

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "123", aliases["id=123"].Id)

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "test", aliases["id=123"].Slug)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "123", aliases["last_read=entity"].Id)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "test", aliases["last_read=entity"].Slug)
}

func TestSavedAliasIsReturnedForANameInLegacyObjectResponse(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"name": "Test Testerson",
		"type": "foo"
	}
}`)

	aliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})

	// Verification

	require.Len(t, aliases, 3, "There should be %d typeToAliasNameToIdMap in map not %d", 3, len(aliases))

	require.Contains(t, aliases, "name=Test_Testerson")
	require.Equal(t, "123", aliases["name=Test_Testerson"].Id)

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "123", aliases["id=123"].Id)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "123", aliases["last_read=entity"].Id)

}

func TestSavedAliasIsReturnedForAnExternalRefInLegacyObjectResponse(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"process_limit": 66,
		"external_ref": "abc-123",
		"type": "foo"
	}
}`)

	aliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})

	// Verification

	require.Len(t, aliases, 3, "There should be %d typeToAliasNameToIdMap in map not %d", 3, len(aliases))

	require.Contains(t, aliases, "external_ref=abc-123")
	require.Equal(t, "123", aliases["external_ref=abc-123"].Id)
	require.Equal(t, "abc-123", aliases["external_ref=abc-123"].ExternalRef)

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "123", aliases["id=123"].Id)
	require.Equal(t, "abc-123", aliases["id=123"].ExternalRef)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "123", aliases["last_read=entity"].Id)
	require.Equal(t, "abc-123", aliases["last_read=entity"].ExternalRef)

}

func TestSavedAliasIsReturnedForAnEmailInComplaintObjectResponse(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo",
		"attributes": {
			"email": "test@test.com"
		}
	}
}`)

	aliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})

	// Verification

	require.Len(t, aliases, 3, "There should be %d typeToAliasNameToIdMap in map not %d", 3, len(aliases))

	require.Contains(t, aliases, "email=test@test.com")
	require.Equal(t, "123", aliases["email=test@test.com"].Id)

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "123", aliases["id=123"].Id)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "123", aliases["last_read=entity"].Id)

}

func TestSavedAliasIsReturnedForASkuInComplaintObjectResponse(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo",
		"attributes": {
			"sku": "test"
		}
	}
}`)

	aliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})

	// Verification

	require.Len(t, aliases, 3, "There should be %d typeToAliasNameToIdMap in map not %d", 3, len(aliases))

	require.Contains(t, aliases, "sku=test")
	require.Equal(t, "123", aliases["sku=test"].Id)

	require.Contains(t, aliases, "sku=test")
	require.Equal(t, "test", aliases["sku=test"].Sku)

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "123", aliases["id=123"].Id)

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "test", aliases["id=123"].Sku)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "123", aliases["last_read=entity"].Id)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "test", aliases["last_read=entity"].Sku)
}

func TestSavedAliasIsReturnedForASlugInComplaintObjectResponse(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo",
		"attributes": {
			"slug": "test"
		}
	}
}`)

	aliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})

	// Verification

	require.Len(t, aliases, 3, "There should be %d typeToAliasNameToIdMap in map not %d", 3, len(aliases))

	require.Contains(t, aliases, "slug=test")
	require.Equal(t, "123", aliases["slug=test"].Id)

	require.Contains(t, aliases, "slug=test")
	require.Equal(t, "test", aliases["slug=test"].Slug)

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "123", aliases["id=123"].Id)

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "test", aliases["id=123"].Slug)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "123", aliases["last_read=entity"].Id)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "test", aliases["last_read=entity"].Slug)
}

func TestSavedAliasIsReturnedForAnExternalRefInComplaintObjectResponse(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo",
		"attributes": {
			"process_limit": 66,
			"external_ref": "abc-123"
		}
	}
}`)

	aliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})

	// Verification

	require.Len(t, aliases, 3, "There should be %d typeToAliasNameToIdMap in map not %d", 3, len(aliases))

	require.Contains(t, aliases, "external_ref=abc-123")
	require.Equal(t, "123", aliases["external_ref=abc-123"].Id)
	require.Equal(t, "abc-123", aliases["external_ref=abc-123"].ExternalRef)

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "123", aliases["id=123"].Id)
	require.Equal(t, "abc-123", aliases["id=123"].ExternalRef)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "123", aliases["last_read=entity"].Id)
	require.Equal(t, "abc-123", aliases["last_read=entity"].ExternalRef)

}

func TestSavedAliasIsReturnedForANameInComplaintObjectResponse(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo",
		"attributes": {
			"name": "Test Testerson"
		}
	}
}`)

	aliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})

	// Verification

	require.Len(t, aliases, 3, "There should be %d typeToAliasNameToIdMap in map not %d", 3, len(aliases))

	require.Contains(t, aliases, "name=Test_Testerson")
	require.Equal(t, "123", aliases["name=Test_Testerson"].Id)

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "123", aliases["id=123"].Id)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "123", aliases["last_read=entity"].Id)

}

func TestSavedAliasIsReturnedForARelationshipObjectInArrayResponse(t *testing.T) {
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data":	[{
		"id": "123",
		"type": "foo",
		"name": "Test Testerson",
		"relationships": {
			"buz": {
				"data": {
					"type": "bar",
					"id": "abc"
				}
			}
		}
	},{
		"id": "456",
		"type": "foo",
		"name": "Bob Robertson",
		"relationships": {
			"buz": {
				"data": {
					"type": "bar",
					"id": "def"
				}
			}
		}
	}]
	
}`)

	aliases := GetAliasesForJsonApiTypeAndAlternates("bar", []string{})

	require.Len(t, aliases, 8, "There should be %d typeToAliasNameToIdMap in map not %d", 8, len(aliases))

	require.Contains(t, aliases, "id=abc")
	require.Equal(t, "abc", aliases["id=abc"].Id)

	require.Contains(t, aliases, "related_buz_for_foo_id=123")
	require.Equal(t, "abc", aliases["related_buz_for_foo_id=123"].Id)

	require.Contains(t, aliases, "related_buz_for_foo_last_read=array[0]")
	require.Equal(t, "abc", aliases["related_buz_for_foo_last_read=array[0]"].Id)

	require.Contains(t, aliases, "related_buz_for_foo_name=Test_Testerson")
	require.Equal(t, "abc", aliases["related_buz_for_foo_name=Test_Testerson"].Id)

	require.Contains(t, aliases, "id=def")
	require.Equal(t, "def", aliases["id=def"].Id)

	require.Contains(t, aliases, "related_buz_for_foo_id=456")
	require.Equal(t, "def", aliases["related_buz_for_foo_id=456"].Id)

	require.Contains(t, aliases, "related_buz_for_foo_last_read=array[1]")
	require.Equal(t, "def", aliases["related_buz_for_foo_last_read=array[1]"].Id)

	require.Contains(t, aliases, "related_buz_for_foo_name=Bob_Robertson")
	require.Equal(t, "def", aliases["related_buz_for_foo_name=Bob_Robertson"].Id)
}

func TestSavedAliasIsReturnedForARelationshipObjectInSingleResponse(t *testing.T) {
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo",
		"name": "Test Testerson",
		"relationships": {
			"buz": {
				"data": {
					"type": "bar",
					"id": "456"
				}
			}
		}
	}
	
}`)

	aliases := GetAliasesForJsonApiTypeAndAlternates("bar", []string{})

	require.Len(t, aliases, 4, "There should be %d typeToAliasNameToIdMap in map not %d", 4, len(aliases))

	require.Contains(t, aliases, "id=456")
	require.Equal(t, "456", aliases["id=456"].Id)

	require.Contains(t, aliases, "related_buz_for_foo_id=123")
	require.Equal(t, "456", aliases["related_buz_for_foo_id=123"].Id)

	require.Contains(t, aliases, "related_buz_for_foo_last_read=entity")
	require.Equal(t, "456", aliases["related_buz_for_foo_last_read=entity"].Id)

	require.Contains(t, aliases, "related_buz_for_foo_name=Test_Testerson")
	require.Equal(t, "456", aliases["related_buz_for_foo_name=Test_Testerson"].Id)
}

func TestResolveAliasValuesReturnsAliasForMatchingValue(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo"
	}
}`)
	value := ResolveAliasValuesOrReturnIdentity("foo", []string{}, "id=123", "id")

	// Verification

	if value != "123" {
		t.Errorf("Alias value of 123 should have been returned, but got %s", value)
		return
	}
}

func TestResolveAliasValuesReturnsAliasSkuForMatchingValue(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo",
		"sku": "mysku"
		
	}
}`)
	value := ResolveAliasValuesOrReturnIdentity("foo", []string{}, "id=123", "sku")

	// Verification

	if value != "mysku" {
		t.Errorf("Alias value of 123 should have been returned, but got %s", value)
		return
	}
}

func TestResolveAliasValuesReturnsAliasCodeForMatchingValue(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo",
		"code": "hello"
		
	}
}`)
	value := ResolveAliasValuesOrReturnIdentity("foo", []string{}, "id=123", "code")

	// Verification

	if value != "hello" {
		t.Errorf("Alias value of hello should have been returned, but got %s", value)
		return
	}
}

func TestResolveAliasValuesReturnsAliasSlugForMatchingValue(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo",
		"slug": "test"
		
	}
}`)
	value := ResolveAliasValuesOrReturnIdentity("foo", []string{}, "id=123", "slug")

	// Verification

	if value != "test" {
		t.Errorf("Alias value of test should have been returned, but got %s", value)
		return
	}
}

func TestResolveAliasValuesReturnsAliasForMatchingValueAsAlternateType(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo"
	}
}`)
	value := ResolveAliasValuesOrReturnIdentity("bar", []string{"zoo", "foo"}, "id=123", "id")

	// Verification

	if value != "123" {
		t.Errorf("Alias value of 123 should have been returned, but got %s", value)
		return
	}
}

func TestResolveAliasValuesReturnsAliasForTypeAndNotAlternateTypeWhenCollisionOccurs(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": [
	{
		"id": "123",
		"name": "hello",
		"type": "foo"
	},
	{
		"id": "456",
		"name": "hello",
		"type": "bar"
	}
]
}`)
	value := ResolveAliasValuesOrReturnIdentity("foo", []string{"zoo", "bar"}, "name=hello", "id")

	// Verification

	if value != "123" {
		t.Errorf("Alias value of 123 should have been returned, but got %s", value)
		return
	}
}

func TestResolveAliasValuesReturnsRequestForUnMatchingValue(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo"
	}
}`)

	value := ResolveAliasValuesOrReturnIdentity("foo", []string{}, "id=ABC", "id")

	// Verification

	if value != "id=ABC" {
		t.Errorf("Alias value of id=ABC should have been returned, but got %s", value)
		return
	}
}

// This test helps prevent crashes from missing directories and some such.
func TestResolveAliasValuesReturnsRequestForUnMatchingValueAndType(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	// Execute SUT
	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo"
	}
}`)

	value := ResolveAliasValuesOrReturnIdentity("bar", []string{}, "id=XYZ", "id")

	// Verification

	if value != "id=XYZ" {
		t.Errorf("Alias value of id=XYZ should have been returned, but got %s", value)
		return
	}
}

func TestClearAllAliasesClearsAllAliases(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo"
	}
}`)

	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "456",
		"type": "bar"
	}
}`)

	// Execute SUT
	err = ClearAllAliases()
	if err != nil {
		t.Errorf("Couldn't clear typeToAliasNameToIdMap %v", err)
		return
	}

	fooAliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})
	barAliases := GetAliasesForJsonApiTypeAndAlternates("bar", []string{})

	// Verification
	if len(fooAliases) != 0 {
		t.Errorf("There should be zero alias for the type foo, not %d", len(fooAliases))
		return
	}

	if len(barAliases) != 0 {
		t.Errorf("There should be zero alias for the type bar, not %d", len(barAliases))
		return
	}

}

func TestClearAllAliasesForJsonTypeOnlyClearsJsonType(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo"
	}
}`)

	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "456",
		"type": "bar"
	}
}`)

	// Execute SUT
	err = ClearAllAliasesForJsonApiType("foo")

	if err != nil {
		t.Errorf("Couldn't clear typeToAliasNameToIdMap %v", err)
		return
	}

	fooAliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})
	barAliases := GetAliasesForJsonApiTypeAndAlternates("bar", []string{})

	// Verification
	if len(fooAliases) != 0 {
		t.Errorf("There should be zero alias for the type foo, not %d", len(fooAliases))
		return
	}

	if len(barAliases) != 2 {
		t.Errorf("There should be two alias for the type bar, not %d", len(barAliases))
		return
	}

}

func TestThatCorruptAliasFileDoesntCrashProgramWhenReadingAliases(t *testing.T) {
	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo"
	}
}`)

	// Execute SUT
	require.Equal(t, 1, FlushAliases(), "Should have written 1 type to disk")

	fileName := getAliasFileForJsonApiType(getAliasDataDirectory(), "foo")

	if err := os.Remove(getAliasFileForJsonApiType(getAliasDataDirectory(), "foo")); err != nil && !os.IsNotExist(err) {
		t.Errorf("Should have been able to delete the file, but got %v ", err)
		return
	}

	err = os.WriteFile(fileName, []byte("{{{"), 0600)
	if err != nil {
		t.Errorf("Couldn't save corrupted yaml file %v", err)
		return
	}

	aliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})

	// Verification
	require.Len(t, aliases, 0, "There should be %d typeToAliasNameToIdMap in map not %d", 0, len(aliases))

}

func TestThatCorruptAliasFileDoesntCrashProgramWhenSavingAliases(t *testing.T) {
	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear typeToAliasNameToIdMap")
	}

	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123",
		"type": "foo"
	}
}`)

	// Execute SUT
	require.Equal(t, 1, FlushAliases(), "Should have written 1 type to disk")
	fileName := getAliasFileForJsonApiType(getAliasDataDirectory(), "foo")

	if err := os.Remove(getAliasFileForJsonApiType(getAliasDataDirectory(), "foo")); err != nil && !os.IsNotExist(err) {
		t.Errorf("Should have been able to delete the file, but got %v ", err)
		return
	}

	err = os.WriteFile(fileName, []byte("{{{"), 0600)
	if err != nil {
		t.Errorf("Couldn't save corrupted yaml file %v", err)
		return
	}

	SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "456",
		"type": "foo"
	}
}`)

	aliases := GetAliasesForJsonApiTypeAndAlternates("foo", []string{})

	// Verification
	require.Len(t, aliases, 2, "There should be %d typeToAliasNameToIdMap in map not %d", 2, len(aliases))

	require.Contains(t, aliases, "id=456")
	require.Equal(t, "456", aliases["id=456"].Id)

}
