package aliases

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"testing"
)

func init() {
	dir, err := ioutil.TempDir("", "epcc-cli-aliases-testing")
	if err != nil {
		log.Panic("Could not create directory", err)
	}

	aliasDirectoryOverride = dir
	log.Infof("Alias directory for tests is %s", dir)
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

	if len(aliases) != 1 {
		t.Errorf("There should be one alias for the type foo, not %d", len(aliases))
	}

	if aliases["id=123"] != nil && aliases["id=123"].Id != "123" {
		t.Errorf("Alias should exist for id=123")
	}
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

	if len(aliases) != 2 {
		t.Errorf("There should be two aliases for the type foo, not %d", len(aliases))
	}

	if aliases["id=123"] != nil && aliases["id=123"].Id != "123" {
		t.Errorf("Alias should exist for id=123")
	}

	if aliases["id=456"] != nil && aliases["id=456"].Id != "456" {
		t.Errorf("Alias should exist for id=456")
	}
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

	if len(aliases) != 3 {
		t.Errorf("There should be three aliases for the type foo, not %d", len(aliases))
	}

	if aliases["id=123"] != nil && aliases["id=123"].Id != "123" {
		t.Errorf("Alias should exist for id=123")
	}

	if aliases["id=456"] != nil && aliases["id=456"].Id != "456" {
		t.Errorf("Alias should exist for id=456")
	}

	if aliases["name=Alpha"] != nil && aliases["name=Alpha"].Id != "456" {
		t.Errorf("Alias should exist for id=456")
	}
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

	if len(aliases) != 2 {
		t.Errorf("There should be two aliases for the type foo, not %d", len(aliases))
	}

	if aliases["id=123"] != nil && aliases["id=123"].Id != "123" {
		t.Errorf("Alias should exist for id=123")
	}

	if aliases["name=Beta"] != nil && aliases["name=Beta"].Id != "123" {
		t.Errorf("Alias should exist for id=123")
	}
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

	if len(aliases) != 1 {
		t.Errorf("There should be one alias for the type foo, not %d", len(aliases))
	}

	if aliases["id=456"] != nil && aliases["id=456"].Id != "456" {
		t.Errorf("Alias should exist for id=456")
	}
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

	if len(aliases) != 2 {
		t.Errorf("There should be one alias for the type foo, not %d", len(aliases))
	}

	if aliases["id=123"] != nil && aliases["id=123"].Id != "123" {
		t.Errorf("Alias should exist for id=123")
	}

	if aliases["id=456"] != nil && aliases["id=456"].Id != "456" {
		t.Errorf("Alias should exist for id=123")
	}
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

	if len(aliases) != 2 {
		t.Errorf("There should be one alias for the type foo, not %d", len(aliases))
	}

	if aliases["email=test@test.com"] != nil && aliases["email=test@test.com"].Id != "123" {
		t.Errorf("Alias should exist for email=test@test.com")
	}

	if aliases["id=123"] != nil && aliases["id=123"].Id != "123" {
		t.Errorf("Alias should exist for id=123")
	}
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

	if len(aliases) != 2 {
		t.Errorf("There should be one alias for the type foo, not %d", len(aliases))
	}

	if aliases["sku=test"] != nil && aliases["sku=test"].Id != "123" {
		t.Errorf("Alias should exist for sku=test")
	}

	if aliases["sku=test"] != nil && aliases["sku=test"].Sku != "test" {
		t.Errorf("Alias should exist for sku=test and have sku test")
	}

	if aliases["id=123"] != nil && aliases["id=123"].Id != "123" {
		t.Errorf("Alias should exist for id=123")
	}

	if aliases["id=123"] != nil && aliases["id=123"].Sku != "test" {
		t.Errorf("Alias should exist for id=123 and have a sku of test")
	}
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

	if len(aliases) != 2 {
		t.Errorf("There should be one alias for the type foo, not %d", len(aliases))
	}

	if aliases["slug=test"] != nil && aliases["slug=test"].Id != "123" {
		t.Errorf("Alias should exist for slug=test")
	}

	if aliases["slug=test"] != nil && aliases["slug=test"].Slug != "test" {
		t.Errorf("Alias should exist for slug=test and have slug value of test")
	}

	if aliases["id=123"] != nil && aliases["id=123"].Id != "123" {
		t.Errorf("Alias should exist for id=123")
	}

	if aliases["id=123"] != nil && aliases["id=123"].Slug != "test" {
		t.Errorf("Alias should exist for id=123 and have a slug of test")
	}
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

	if len(aliases) != 2 {
		t.Errorf("There should be one alias for the type foo, not %d", len(aliases))
	}

	if aliases["name=Test_Testerson"] != nil && aliases["name=Test_Testerson"].Id != "123" {
		t.Errorf("Alias should exist for name=Test_Testerson")
	}

	if aliases["id=123"] != nil && aliases["id=123"].Id != "123" {
		t.Errorf("Alias should exist for id=123")
	}
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

	if len(aliases) != 2 {
		t.Errorf("There should be one alias for the type foo, not %d", len(aliases))
	}

	if aliases["email=test@test.com"] != nil && aliases["email=test@test.com"].Id != "123" {
		t.Errorf("Alias should exist for email=test@test.com")
	}

	if aliases["id=123"] != nil && aliases["id=123"].Id != "123" {
		t.Errorf("Alias should exist for id=123")
	}
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

	if len(aliases) != 2 {
		t.Errorf("There should be one alias for the type foo, not %d", len(aliases))
	}

	if aliases["sku=test"] != nil && aliases["sku=test"].Id != "123" {
		t.Errorf("Alias should exist for sku=test")
	}

	if aliases["sku=test"] != nil && aliases["sku=test"].Sku != "test" {
		t.Errorf("Alias should exist for sku=test and have a sku of test")
	}

	if aliases["id=123"] != nil && aliases["id=123"].Id != "123" {
		t.Errorf("Alias should exist for id=123")
	}

	if aliases["id=123"] != nil && aliases["id=123"].Sku != "test" {
		t.Errorf("Alias should exist for id=123 and have a sku of test")
	}
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

	if len(aliases) != 2 {
		t.Errorf("There should be one alias for the type foo, not %d", len(aliases))
	}

	if aliases["slug=test"] != nil && aliases["slug=test"].Id != "123" {
		t.Errorf("Alias should exist for slug=test")
	}

	if aliases["slug=test"] != nil && aliases["slug=test"].Slug != "test" {
		t.Errorf("Alias should exist for slug=test and have a slug of test")
	}

	if aliases["id=123"] != nil && aliases["id=123"].Id != "123" {
		t.Errorf("Alias should exist for id=123")
	}

	if aliases["id=123"] != nil && aliases["id=123"].Slug != "test" {
		t.Errorf("Alias should exist for id=123 and have a slug of test")
	}
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

	if len(aliases) != 2 {
		t.Errorf("There should be one alias for the type foo, not %d", len(aliases))
	}

	if aliases["name=Test_Testerson"] != nil && aliases["name=Test_Testerson"].Id != "123" {
		t.Errorf("Alias should exist for name=Test_Testerson")
	}

	if aliases["id=123"] != nil && aliases["id=123"].Id != "123" {
		t.Errorf("Alias should exist for id=123")
	}
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
	value := ResolveAliasValuesOrReturnIdentity("foo", "id=123")

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

	value := ResolveAliasValuesOrReturnIdentity("foo", "id=ABC")

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

	value := ResolveAliasValuesOrReturnIdentity("bar", "id=XYZ")

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

	if len(barAliases) != 1 {
		t.Errorf("There should be one alias for the type bar, not %d", len(barAliases))
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
	if len(aliases) != 1 {
		t.Errorf("There should be one alias for the type foo, not %d", len(aliases))
	}

	if aliases["id=456"] != nil && aliases["id=456"].Id != "456" {
		t.Errorf("Alias should exist for id=456")
	}

}
