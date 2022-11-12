package aliases

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func init() {
	InitializeAliasDirectoryForTesting()
}

func TestSavedAliasIsReturnedInAllAliasesForSingleResponse(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear aliases")
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

	aliases := GetAliasesForJsonApiType("foo")

	// Verification
	require.Len(t, aliases, 2, "There should be %d aliases in map not %d", 2, len(aliases))

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "123", aliases["id=123"].Id)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "123", aliases["last_read=entity"].Id)
}

func TestSavedAliasAppendsAndPreservesPreviousUnrelatedAliases(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear aliases")
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

	aliases := GetAliasesForJsonApiType("foo")

	// Verification

	require.Len(t, aliases, 3, "There should be %d aliases in map not %d", 3, len(aliases))

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
		t.Fatalf("Could not clear aliases")
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

	aliases := GetAliasesForJsonApiType("foo")

	// Verification

	require.Len(t, aliases, 4, "There should be %d aliases in map not %d", 4, len(aliases))

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
		t.Fatalf("Could not clear aliases")
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

	aliases := GetAliasesForJsonApiType("foo")

	// Verification

	require.Len(t, aliases, 3, "There should be %d aliases in map not %d", 3, len(aliases))

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "123", aliases["id=123"].Id)

	require.Contains(t, aliases, "name=Beta")
	require.Equal(t, "123", aliases["name=Beta"].Id)
	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "123", aliases["last_read=entity"].Id)
}

func TestDeleteAliasByIdDeletesAnAlias(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear aliases")
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

	aliases := GetAliasesForJsonApiType("foo")

	// Verification

	require.Len(t, aliases, 2, "There should be %d aliases in map not %d", 2, len(aliases))

	require.Contains(t, aliases, "id=456")
	require.Equal(t, "456", aliases["id=456"].Id)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "456", aliases["last_read=entity"].Id)
}

func TestAllAliasesAreReturnedInAllAliasesForArrayResponse(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear aliases")
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

	aliases := GetAliasesForJsonApiType("foo")

	// Verification

	require.Len(t, aliases, 4, "There should be %d aliases in map not %d", 4, len(aliases))

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
		t.Fatalf("Could not clear aliases")
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

	aliases := GetAliasesForJsonApiType("foo")

	// Verification

	require.Len(t, aliases, 3, "There should be %d aliases in map not %d", 3, len(aliases))

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
		t.Fatalf("Could not clear aliases")
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

	aliases := GetAliasesForJsonApiType("foo")

	// Verification

	require.Len(t, aliases, 3, "There should be %d aliases in map not %d", 3, len(aliases))

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

func TestSavedAliasIsReturnedForASlugInLegacyObjectResponse(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear aliases")
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

	aliases := GetAliasesForJsonApiType("foo")

	// Verification

	require.Len(t, aliases, 3, "There should be %d aliases in map not %d", 3, len(aliases))

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
		t.Fatalf("Could not clear aliases")
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

	aliases := GetAliasesForJsonApiType("foo")

	// Verification

	require.Len(t, aliases, 3, "There should be %d aliases in map not %d", 3, len(aliases))

	require.Contains(t, aliases, "name=Test_Testerson")
	require.Equal(t, "123", aliases["name=Test_Testerson"].Id)

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "123", aliases["id=123"].Id)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "123", aliases["last_read=entity"].Id)

}

func TestSavedAliasIsReturnedForAnEmailInComplaintObjectResponse(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear aliases")
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

	aliases := GetAliasesForJsonApiType("foo")

	// Verification

	require.Len(t, aliases, 3, "There should be %d aliases in map not %d", 3, len(aliases))

	require.Contains(t, aliases, "email=test@test.com")
	require.Equal(t, "123", aliases["email=test@test.com"].Id)

	require.Contains(t, aliases, "id=123")
	require.Equal(t, "123", aliases["id=123"].Id)

	require.Contains(t, aliases, "last_read=entity")
	require.Equal(t, "123", aliases["last_read=entity"].Id)

}

func TestSavedAliasIsReturnedForAnSkuInComplaintObjectResponse(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear aliases")
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

	aliases := GetAliasesForJsonApiType("foo")

	// Verification

	require.Len(t, aliases, 3, "There should be %d aliases in map not %d", 3, len(aliases))

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
		t.Fatalf("Could not clear aliases")
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

	aliases := GetAliasesForJsonApiType("foo")

	// Verification

	require.Len(t, aliases, 3, "There should be %d aliases in map not %d", 3, len(aliases))

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

func TestSavedAliasIsReturnedForANameInComplaintObjectResponse(t *testing.T) {

	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear aliases")
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

	aliases := GetAliasesForJsonApiType("foo")

	// Verification

	require.Len(t, aliases, 3, "There should be %d aliases in map not %d", 3, len(aliases))

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
		t.Fatalf("Could not clear aliases")
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

	aliases := GetAliasesForJsonApiType("bar")

	require.Len(t, aliases, 8, "There should be %d aliases in map not %d", 8, len(aliases))

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
		t.Fatalf("Could not clear aliases")
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

	aliases := GetAliasesForJsonApiType("bar")

	require.Len(t, aliases, 4, "There should be %d aliases in map not %d", 4, len(aliases))

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
		t.Fatalf("Could not clear aliases")
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
	value := ResolveAliasValuesOrReturnIdentity("foo", "id=123", "id")

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
		t.Fatalf("Could not clear aliases")
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

	value := ResolveAliasValuesOrReturnIdentity("foo", "id=ABC", "id")

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
		t.Fatalf("Could not clear aliases")
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

	value := ResolveAliasValuesOrReturnIdentity("bar", "id=XYZ", "id")

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
		t.Fatalf("Could not clear aliases")
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
		t.Errorf("Couldn't clear aliases %v", err)
		return
	}

	fooAliases := GetAliasesForJsonApiType("foo")
	barAliases := GetAliasesForJsonApiType("bar")

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
		t.Fatalf("Could not clear aliases")
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
		t.Errorf("Couldn't clear aliases %v", err)
		return
	}

	fooAliases := GetAliasesForJsonApiType("foo")
	barAliases := GetAliasesForJsonApiType("bar")

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
		t.Fatalf("Could not clear aliases")
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

	aliases := GetAliasesForJsonApiType("foo")

	// Verification
	require.Len(t, aliases, 0, "There should be %d aliases in map not %d", 0, len(aliases))

}

func TestThatCorruptAliasFileDoesntCrashProgramWhenSavingAliases(t *testing.T) {
	// Fixture Setup
	err := ClearAllAliases()
	if err != nil {
		t.Fatalf("Could not clear aliases")
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

	aliases := GetAliasesForJsonApiType("foo")

	// Verification
	require.Len(t, aliases, 2, "There should be %d aliases in map not %d", 2, len(aliases))

	require.Contains(t, aliases, "id=456")
	require.Equal(t, "456", aliases["id=456"].Id)

}
