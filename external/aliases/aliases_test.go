package aliases

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
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
	if len(aliases) != 2 {
		t.Errorf("There should be one alias for the type foo, not %d", len(aliases))
	}

	assert.Contains(t, aliases, "id=123")
	assert.Equal(t, "123", aliases["id=123"].Id)

	assert.Contains(t, aliases, "last_read=entity")
	assert.Equal(t, "123", aliases["last_read=entity"].Id)
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

	if len(aliases) != 3 {
		t.Errorf("There should be two aliases for the type foo, not %d", len(aliases))
	}

	assert.Contains(t, aliases, "id=123")
	assert.Equal(t, "123", aliases["id=123"].Id)

	assert.Contains(t, aliases, "id=456")
	assert.Equal(t, "456", aliases["id=456"].Id)

	assert.Contains(t, aliases, "last_read=entity")
	assert.Equal(t, "456", aliases["last_read=entity"].Id)
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

	if len(aliases) != 4 {
		t.Errorf("There should be three aliases for the type foo, not %d", len(aliases))
	}

	assert.Contains(t, aliases, "id=123")
	assert.Equal(t, "123", aliases["id=123"].Id)

	assert.Contains(t, aliases, "id=456")
	assert.Equal(t, "456", aliases["id=456"].Id)

	assert.Contains(t, aliases, "name=Alpha")
	assert.Equal(t, "456", aliases["name=Alpha"].Id)

	assert.Contains(t, aliases, "last_read=entity")
	assert.Equal(t, "456", aliases["last_read=entity"].Id)
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

	if len(aliases) != 3 {
		t.Errorf("There should be three aliases for the type foo, not %d", len(aliases))
	}

	assert.Contains(t, aliases, "id=123")
	assert.Equal(t, "123", aliases["id=123"].Id)

	assert.Contains(t, aliases, "name=Beta")
	assert.Equal(t, "123", aliases["name=Beta"].Id)
	assert.Contains(t, aliases, "last_read=entity")
	assert.Equal(t, "123", aliases["last_read=entity"].Id)
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

	if len(aliases) != 2 {
		t.Errorf("There should be two alias for the type foo, not %d", len(aliases))
	}

	assert.Contains(t, aliases, "id=456")
	assert.Equal(t, "456", aliases["id=456"].Id)

	assert.Contains(t, aliases, "last_read=entity")
	assert.Equal(t, "456", aliases["last_read=entity"].Id)
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

	if len(aliases) != 4 {
		t.Errorf("There should be four aliases for the type foo, not %d", len(aliases))
	}

	assert.Contains(t, aliases, "id=123")
	assert.Equal(t, "123", aliases["id=123"].Id)

	assert.Contains(t, aliases, "id=456")
	assert.Equal(t, "456", aliases["id=456"].Id)

	assert.Contains(t, aliases, "last_read=array[0]")
	assert.Equal(t, "123", aliases["last_read=array[0]"].Id)

	assert.Contains(t, aliases, "last_read=array[1]")
	assert.Equal(t, "456", aliases["last_read=array[1]"].Id)
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

	if len(aliases) != 3 {
		t.Errorf("There should be three alias for the type foo, not %d", len(aliases))
	}

	assert.Contains(t, aliases, "email=test@test.com")
	assert.Equal(t, "123", aliases["email=test@test.com"].Id)

	assert.Contains(t, aliases, "id=123")
	assert.Equal(t, "123", aliases["id=123"].Id)

	assert.Contains(t, aliases, "last_read=entity")
	assert.Equal(t, "123", aliases["last_read=entity"].Id)
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

	if len(aliases) != 3 {
		t.Errorf("There should be three aliases for the type foo, not %d", len(aliases))
	}

	assert.Contains(t, aliases, "sku=test")
	assert.Equal(t, "123", aliases["sku=test"].Id)

	assert.Contains(t, aliases, "sku=test")
	assert.Equal(t, "test", aliases["sku=test"].Sku)

	assert.Contains(t, aliases, "id=123")
	assert.Equal(t, "123", aliases["id=123"].Id)

	assert.Contains(t, aliases, "id=123")
	assert.Equal(t, "test", aliases["id=123"].Sku)

	assert.Contains(t, aliases, "last_read=entity")
	assert.Equal(t, "123", aliases["last_read=entity"].Id)

	assert.Contains(t, aliases, "last_read=entity")
	assert.Equal(t, "test", aliases["last_read=entity"].Sku)
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

	if len(aliases) != 3 {
		t.Errorf("There should be three aliases for the type foo, not %d", len(aliases))
	}

	assert.Contains(t, aliases, "slug=test")
	assert.Equal(t, "123", aliases["slug=test"].Id)

	assert.Contains(t, aliases, "slug=test")
	assert.Equal(t, "test", aliases["slug=test"].Slug)

	assert.Contains(t, aliases, "id=123")
	assert.Equal(t, "123", aliases["id=123"].Id)

	assert.Contains(t, aliases, "id=123")
	assert.Equal(t, "test", aliases["id=123"].Slug)

	assert.Contains(t, aliases, "last_read=entity")
	assert.Equal(t, "123", aliases["last_read=entity"].Id)

	assert.Contains(t, aliases, "last_read=entity")
	assert.Equal(t, "test", aliases["last_read=entity"].Slug)
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

	if len(aliases) != 3 {
		t.Errorf("There should be three aliases for the type foo, not %d", len(aliases))
	}

	assert.Contains(t, aliases, "name=Test_Testerson")
	assert.Equal(t, "123", aliases["name=Test_Testerson"].Id)

	assert.Contains(t, aliases, "id=123")
	assert.Equal(t, "123", aliases["id=123"].Id)

	assert.Contains(t, aliases, "last_read=entity")
	assert.Equal(t, "123", aliases["last_read=entity"].Id)

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

	if len(aliases) != 3 {
		t.Errorf("There should be three aliases for the type foo, not %d", len(aliases))
	}

	assert.Contains(t, aliases, "email=test@test.com")
	assert.Equal(t, "123", aliases["email=test@test.com"].Id)

	assert.Contains(t, aliases, "id=123")
	assert.Equal(t, "123", aliases["id=123"].Id)

	assert.Contains(t, aliases, "last_read=entity")
	assert.Equal(t, "123", aliases["last_read=entity"].Id)

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

	if len(aliases) != 3 {
		t.Errorf("There should be three alias for the type foo, not %d", len(aliases))
	}

	assert.Contains(t, aliases, "sku=test")
	assert.Equal(t, "123", aliases["sku=test"].Id)

	assert.Contains(t, aliases, "sku=test")
	assert.Equal(t, "test", aliases["sku=test"].Sku)

	assert.Contains(t, aliases, "id=123")
	assert.Equal(t, "123", aliases["id=123"].Id)

	assert.Contains(t, aliases, "id=123")
	assert.Equal(t, "test", aliases["id=123"].Sku)

	assert.Contains(t, aliases, "last_read=entity")
	assert.Equal(t, "123", aliases["last_read=entity"].Id)

	assert.Contains(t, aliases, "last_read=entity")
	assert.Equal(t, "test", aliases["last_read=entity"].Sku)
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

	if len(aliases) != 3 {
		t.Errorf("There should be three aliases for the type foo, not %d", len(aliases))
	}

	assert.Contains(t, aliases, "slug=test")
	assert.Equal(t, "123", aliases["slug=test"].Id)

	assert.Contains(t, aliases, "slug=test")
	assert.Equal(t, "test", aliases["slug=test"].Slug)

	assert.Contains(t, aliases, "id=123")
	assert.Equal(t, "123", aliases["id=123"].Id)

	assert.Contains(t, aliases, "id=123")
	assert.Equal(t, "test", aliases["id=123"].Slug)

	assert.Contains(t, aliases, "last_read=entity")
	assert.Equal(t, "123", aliases["last_read=entity"].Id)

	assert.Contains(t, aliases, "last_read=entity")
	assert.Equal(t, "test", aliases["last_read=entity"].Slug)
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

	if len(aliases) != 3 {
		t.Errorf("There should be three alias for the type foo, not %d", len(aliases))
	}

	assert.Contains(t, aliases, "name=Test_Testerson")
	assert.Equal(t, "123", aliases["name=Test_Testerson"].Id)

	assert.Contains(t, aliases, "id=123")
	assert.Equal(t, "123", aliases["id=123"].Id)

	assert.Contains(t, aliases, "last_read=entity")
	assert.Equal(t, "123", aliases["last_read=entity"].Id)

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

	if len(aliases) != 8 {
		t.Errorf("There should be eight aliases for the type bar, not %d", len(aliases))
	}

	assert.Contains(t, aliases, "id=abc")
	assert.Equal(t, "abc", aliases["id=abc"].Id)

	assert.Contains(t, aliases, "related_buz_for_foo_id=123")
	assert.Equal(t, "abc", aliases["related_buz_for_foo_id=123"].Id)

	assert.Contains(t, aliases, "related_buz_for_foo_last_read=array[0]")
	assert.Equal(t, "abc", aliases["related_buz_for_foo_last_read=array[0]"].Id)

	assert.Contains(t, aliases, "related_buz_for_foo_name=Test_Testerson")
	assert.Equal(t, "abc", aliases["related_buz_for_foo_name=Test_Testerson"].Id)

	assert.Contains(t, aliases, "id=def")
	assert.Equal(t, "def", aliases["id=def"].Id)

	assert.Contains(t, aliases, "related_buz_for_foo_id=456")
	assert.Equal(t, "def", aliases["related_buz_for_foo_id=456"].Id)

	assert.Contains(t, aliases, "related_buz_for_foo_last_read=array[1]")
	assert.Equal(t, "def", aliases["related_buz_for_foo_last_read=array[1]"].Id)

	assert.Contains(t, aliases, "related_buz_for_foo_name=Bob_Robertson")
	assert.Equal(t, "def", aliases["related_buz_for_foo_name=Bob_Robertson"].Id)
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

	if len(aliases) != 4 {
		t.Errorf("There should be four aliases for the type bar, not %d", len(aliases))
	}

	assert.Contains(t, aliases, "id=456")
	assert.Equal(t, "456", aliases["id=456"].Id)

	assert.Contains(t, aliases, "related_buz_for_foo_id=123")
	assert.Equal(t, "456", aliases["related_buz_for_foo_id=123"].Id)

	assert.Contains(t, aliases, "related_buz_for_foo_last_read=entity")
	assert.Equal(t, "456", aliases["related_buz_for_foo_last_read=entity"].Id)

	assert.Contains(t, aliases, "related_buz_for_foo_name=Test_Testerson")
	assert.Equal(t, "456", aliases["related_buz_for_foo_name=Test_Testerson"].Id)
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
	}

	fooAliases := GetAliasesForJsonApiType("foo")
	barAliases := GetAliasesForJsonApiType("bar")

	// Verification
	if len(fooAliases) != 0 {
		t.Errorf("There should be zero alias for the type foo, not %d", len(fooAliases))
	}

	if len(barAliases) != 0 {
		t.Errorf("There should be zero alias for the type bar, not %d", len(barAliases))
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
	}

	fooAliases := GetAliasesForJsonApiType("foo")
	barAliases := GetAliasesForJsonApiType("bar")

	// Verification
	if len(fooAliases) != 0 {
		t.Errorf("There should be zero alias for the type foo, not %d", len(fooAliases))
	}

	if len(barAliases) != 2 {
		t.Errorf("There should be two alias for the type bar, not %d", len(barAliases))
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
	fileName := getAliasFileForJsonApiType(getAliasDataDirectory(), "foo")

	if err := os.Remove(getAliasFileForJsonApiType(getAliasDataDirectory(), "foo")); err != nil && !os.IsNotExist(err) {
		t.Errorf("Should have been able to delete the file, but got %v ", err)
	}

	err = ioutil.WriteFile(fileName, []byte("{{{"), 0600)
	if err != nil {
		t.Errorf("Couldn't save corrupted yaml file %v", err)
	}

	aliases := GetAliasesForJsonApiType("foo")

	// Verification
	if len(aliases) != 0 {
		t.Errorf("There should be zero alias for the type foo, not %d", len(aliases))
	}

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
	fileName := getAliasFileForJsonApiType(getAliasDataDirectory(), "foo")

	if err := os.Remove(getAliasFileForJsonApiType(getAliasDataDirectory(), "foo")); err != nil && !os.IsNotExist(err) {
		t.Errorf("Should have been able to delete the file, but got %v ", err)
	}

	err = ioutil.WriteFile(fileName, []byte("{{{"), 0600)
	if err != nil {
		t.Errorf("Couldn't save corrupted yaml file %v", err)
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
	if len(aliases) != 2 {
		t.Errorf("There should be two aliases for the type foo, not %d", len(aliases))
	}

	assert.Contains(t, aliases, "id=456")
	assert.Equal(t, "456", aliases["id=456"].Id)

}
