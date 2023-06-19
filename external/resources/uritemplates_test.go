package resources

import (
	"github.com/elasticpath/epcc-cli/external/aliases"
	"testing"
)

func init() {
	aliases.InitializeAliasDirectoryForTesting()
}

func TestGetNumberOfVariablesReturnsErrorOnTemplate(t *testing.T) {

	// Fixture Setup
	url := "/v2/{te"

	// Execute SUT
	_, err := GetNumberOfVariablesNeeded(url)

	// Verification
	if err == nil {
		t.Errorf("An invalid uri template should have given us an error, not nil ")
	}

}

func TestGetNumberOfVariablesNeededIsZeroWhenNoVariablesNeeded(t *testing.T) {

	// Fixture Setup
	url := "/v2/flows"

	// Execute SUT
	numberOfVariablesNeeded, _ := GetNumberOfVariablesNeeded(url)

	// Verification

	if numberOfVariablesNeeded != 0 {
		t.Errorf("Expected that the number of variables needed was 0, but got %d", numberOfVariablesNeeded)
	}

}

func TestGetNumberOfVariablesNeededIsOneWhenOneVariablesNeeded(t *testing.T) {

	// Fixture Setup
	url := "/v2/flows/{flows}"

	// Execute SUT
	numberOfVariablesNeeded, _ := GetNumberOfVariablesNeeded(url)

	// Verification

	if numberOfVariablesNeeded != 1 {
		t.Errorf("Expected that the number of variables needed was 1, but got %d", numberOfVariablesNeeded)
	}

}

func TestGetNumberOfVariablesNeededIsThreeWhenThreeVariablesNeeded(t *testing.T) {

	// Fixture Setup
	url := "/v2/flows/{flows}/{accounts}/{customers}"

	// Execute SUT
	numberOfVariablesNeeded, _ := GetNumberOfVariablesNeeded(url)

	// Verification

	if numberOfVariablesNeeded != 3 {
		t.Errorf("Expected that the number of variables needed was 3, but got %d", numberOfVariablesNeeded)
	}

}

func TestGetTypesOfVariablesNeededReturnsErrorWithInvalidUriTemplate(t *testing.T) {
	// Fixture Setup
	url := "/v2/{tes"

	// Execute SUT
	_, err := GetTypesOfVariablesNeeded(url)

	// Verification
	if err == nil {
		t.Errorf("An invalid uri template should have given us an error, not nil ")
	}
}

func TestGetTypesOfVariablesNeededReturnsEmptyArrayWhenNoArguments(t *testing.T) {
	// Fixture Setup
	url := "/v2/customers"

	// Execute SUT
	types, err := GetTypesOfVariablesNeeded(url)

	// Verification
	if err != nil {
		t.Errorf("We should not have gotten an error in this case :(, but got %v", err)
	}

	if len(types) != 0 {
		t.Errorf("Expected the number of types returned is 0, but got %d", len(types))
	}

}

func TestGetTypesOfVariablesNeededReturnsTypeInBaseCase(t *testing.T) {
	// Fixture Setup
	url := "/v2/{customers}"

	// Execute SUT
	types, err := GetTypesOfVariablesNeeded(url)

	// Verification
	if err != nil {
		t.Errorf("We should not have gotten an error in this case :(, but got %v", err)
	}

	if len(types) != 1 {
		t.Errorf("Expected the number of types returned is 1, but got %d", len(types))
	}

	if types[0] != "customers" {
		t.Errorf("Expected that the type of the first argument is customers, not %s", types[0])
	}
}

func TestGetTypesOfVariablesNeededReturnsTypeWithThreeVariables(t *testing.T) {
	// Fixture Setup
	url := "/v2/{customers}/addresses/{flows}/flows/{entries}"

	// Execute SUT
	types, err := GetTypesOfVariablesNeeded(url)

	// Verification
	if err != nil {
		t.Errorf("We should not have gotten an error in this case :(, but got %v", err)
	}

	if len(types) != 3 {
		t.Errorf("Expected the number of types returned is 3, but got %d", len(types))
	}

	if types[0] != "customers" {
		t.Errorf("Expected that the type of the first argument is customers, not %s", types[0])
	}

	if types[1] != "flows" {
		t.Errorf("Expected that the type of the second argument is flows, not %s", types[1])
	}

	if types[2] != "entries" {
		t.Errorf("Expected that the type of the third argument is entries, not %s", types[2])
	}

}

func TestUriTemplatesTypeConversionConvertsUnderscoresToDashes(t *testing.T) {
	// Fixture Setup
	url := "/v2/customers/{customers}/addresses/{customer_addresses}"

	// Execute SUT
	types, err := GetTypesOfVariablesNeeded(url)

	// Verification
	if err != nil {
		t.Errorf("We should not have gotten an error in this case :(, but got %v", err)
	}

	if len(types) != 2 {
		t.Errorf("Expected the number of types returned is 2, but got %d", len(types))
	}

	if types[0] != "customers" {
		t.Errorf("Expected that the type of the first argument is customers, not %s", types[0])
	}

	if types[1] != "customer-addresses" {
		t.Errorf("Expected that the type of the second argument is customer_addresses, not %s", types[1])
	}

}

func TestGenerateUrlHappyPathWithSlugParentResourceValueOverride(t *testing.T) {
	// Fixture Setup

	err := aliases.ClearAllAliases()
	if err != nil {
		t.Errorf("Couldn't create test fixtures, error while cleaning aliases, %v", err)
	}

	crudEntityInfo := getValidCrudEntityInfo()
	crudEntityInfo.Url = "/v2/flows/{flows}"
	crudEntityInfo.ParentResourceValueOverrides = map[string]string{
		"flows": "slug",
	}

	flowExample := `{
	"data": {
		"id": "123",
		"type": "flow",
		"slug": "test"
	}
}`

	aliases.SaveAliasesForResources(flowExample)

	expectedUrlWithSlugNotId := "/v2/flows/test"

	// Execute SUT

	actualUrl, err := GenerateUrl(&crudEntityInfo, []string{"slug=test"}, true)

	// Verification

	if err != nil {
		t.Errorf("Should not have gotten error when generating URL.")
	}

	if actualUrl != expectedUrlWithSlugNotId {
		t.Errorf("Url should have been %s but got %s", expectedUrlWithSlugNotId, actualUrl)
	}
}

func TestGenerateUrlHappyPathWithNoParentResourceValueOverride(t *testing.T) {
	// Fixture Setup

	err := aliases.ClearAllAliases()
	if err != nil {
		t.Errorf("Couldn't create test fixtures, error while cleaning aliases, %v", err)
	}

	crudEntityInfo := getValidCrudEntityInfo()
	crudEntityInfo.Url = "/v2/customers/{customers}"
	crudEntityInfo.ParentResourceValueOverrides = map[string]string{}

	flowExample := `{
	"data": {
		"id": "123",
		"type": "customer",
		"name": "Ron Swanson"
	}
}`

	aliases.SaveAliasesForResources(flowExample)

	expectedUrlWithId := "/v2/customers/123"

	// Execute SUT

	actualUrl, err := GenerateUrl(&crudEntityInfo, []string{"name=Ron_Swanson"}, true)

	// Verification

	if err != nil {
		t.Errorf("Should not have gotten error when generating URL.")
	}

	if actualUrl != expectedUrlWithId {
		t.Errorf("Url should have been %s but got %s", expectedUrlWithId, actualUrl)
	}
}

func TestGenerateUrlHappyPathWithNoParentResourceValueOverrideAndNoAliasSubstitution(t *testing.T) {
	// Fixture Setup

	err := aliases.ClearAllAliases()
	if err != nil {
		t.Errorf("Couldn't create test fixtures, error while cleaning aliases, %v", err)
	}

	crudEntityInfo := getValidCrudEntityInfo()
	crudEntityInfo.Url = "/v2/customers/{customers}"
	crudEntityInfo.ParentResourceValueOverrides = map[string]string{}

	flowExample := `{
	"data": {
		"id": "123",
		"type": "customer",
		"name": "Ron Swanson"
	}
}`

	aliases.SaveAliasesForResources(flowExample)

	expectedUrlWithId := "/v2/customers/name=Ron_Swanson"

	// Execute SUT

	actualUrl, err := GenerateUrl(&crudEntityInfo, []string{"name=Ron_Swanson"}, false)

	// Verification

	if err != nil {
		t.Errorf("Should not have gotten error when generating URL.")
	}

	if actualUrl != expectedUrlWithId {
		t.Errorf("Url should have been %s but got %s", expectedUrlWithId, actualUrl)
	}
}

func getValidCrudEntityInfo() CrudEntityInfo {
	return CrudEntityInfo{
		Docs:            "https://www.google.ca",
		Url:             "/v2/flows/{flows}",
		ContentType:     "application/json",
		QueryParameters: "",
		MinResources:    0,
	}
}
