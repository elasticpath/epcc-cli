package json

import (
	"fmt"
	"github.com/elasticpath/epcc-cli/external/aliases"
	"github.com/elasticpath/epcc-cli/external/resources"
	"github.com/stretchr/testify/require"
	"testing"
)

func init() {
	aliases.InitializeAliasDirectoryForTesting()
	resources.PublicInit()
}

func TestErrorMessageWhenOddNumberOfValuesPassed(t *testing.T) {
	// Fixture Setup
	input := []string{"[0]"}
	expected := fmt.Errorf("the number of arguments 1 supplied isn't even, json should be passed in key value pairs. Do you have an extra/missing id?")

	// Execute SUT
	_, actual := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual.Error() != expected.Error() {
		t.Fatalf("Testing json conversion of value '%s' did not match\nExpected: %s\nActually: %s", input, expected, actual)
	}
}

func TestToJsonLegacyFormatEmptyValue(t *testing.T) {
	// Fixture Setup
	input := []string{}
	expected := `{"data":{}}`

	// Execute SUT
	actual, _ := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonLegacyFormatSimpleKeyStringValue(t *testing.T) {
	// Fixture Setup
	input := []string{"key", "val"}
	expected := `{"data":{"key":"val"}}`

	// Execute SUT
	actual, _ := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonLegacyFormatSimpleKeyStringValueWithQuotes(t *testing.T) {
	// Fixture Setup
	input := []string{"key", "val\"ue"}
	expected := `{"data":{"key":"val\"ue"}}`

	// Execute SUT
	actual, err := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	require.NoError(t, err)
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonLegacyFormatSimpleKeyStringValueWithNewLines(t *testing.T) {
	// Fixture Setup
	input := []string{"key", "val\nue"}
	expected := `{"data":{"key":"val\nue"}}`

	// Execute SUT
	actual, err := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	require.NoError(t, err)
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonLegacyFormatSimpleKeyStringValueWithBackslashes(t *testing.T) {
	// Fixture Setup
	input := []string{"key", "val\\nue"}
	expected := `{"data":{"key":"val\\nue"}}`

	// Execute SUT
	actual, err := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	require.NoError(t, err)
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonLegacyFormatSimpleNestedKeyValue(t *testing.T) {
	// Fixture Setup
	input := []string{"foo.bar", "val"}
	expected := `{"data":{"foo":{"bar":"val"}}}`

	// Execute SUT
	actual, _ := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonLegacyFormatSimpleKeyNumericValue(t *testing.T) {
	// Fixture Setup
	input := []string{"key", "3"}
	expected := `{"data":{"key":3}}`

	// Execute SUT
	actual, _ := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonLegacyFormatSimpleKeyFloatNumericValue(t *testing.T) {
	// Fixture Setup
	input := []string{"key", "3.3"}
	expected := `{"data":{"key":3.3}}`

	// Execute SUT
	actual, _ := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonLegacyFormatSimpleKeyBooleanTrueValue(t *testing.T) {
	// Fixture Setup
	input := []string{"key", "true"}
	expected := `{"data":{"key":true}}`

	// Execute SUT
	actual, _ := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonLegacyFormatSimpleKeyBooleanFalseValue(t *testing.T) {
	// Fixture Setup
	input := []string{"key", "false"}
	expected := `{"data":{"key":false}}`

	// Execute SUT
	actual, _ := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonLegacyFormatSimpleKeyNullValue(t *testing.T) {
	// Fixture Setup
	input := []string{"key", "null"}
	expected := `{"data":{"key":null}}`

	// Execute SUT
	actual, _ := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonLegacyFormatSimpleKeyEmptyArrayValue(t *testing.T) {
	// Fixture Setup
	input := []string{"key", "[]"}
	expected := `{"data":{"key":[]}}`

	// Execute SUT
	actual, _ := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}

}

func TestToJsonLegacyFormatSimpleArrayIndexValue(t *testing.T) {
	// Fixture Setup
	input := []string{"key[0]", "val"}
	expected := `{"data":{"key":["val"]}}`

	// Execute SUT
	actual, _ := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonLegacyFormatSimpleArrayWithTwoValues(t *testing.T) {
	// Fixture Setup
	input := []string{"key[0]", "val", "key[1]", "val2"}
	expected := `{"data":{"key":["val","val2"]}}`

	// Execute SUT
	actual, _ := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonCompliantFormatEmptyValue(t *testing.T) {
	// Fixture Setup
	input := []string{}
	expected := `{"data":{}}`

	// Execute SUT
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonCompliantFormatSimpleKeyStringValue(t *testing.T) {
	// Fixture Setup
	input := []string{"key", "val"}
	expected := `{"data":{"attributes":{"key":"val"}}}`

	// Execute SUT
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonCompliantFormatSimpleKeyStringValueAttributesKeyNotDoubleEncoded(t *testing.T) {
	// Fixture Setup
	input := []string{"attributes.key", "val"}
	expected := `{"data":{"attributes":{"key":"val"}}}`

	// Execute SUT
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonCompliantFormatSimpleKeyStringValueRelationshipsKeyNotDoubleEncoded(t *testing.T) {
	// Fixture Setup
	input := []string{"relationships.key", "val"}
	expected := `{"data":{"relationships":{"key":"val"}}}`

	// Execute SUT
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonCompliantFormatSimpleKeyWithTypeStringValue(t *testing.T) {
	// Fixture Setup
	input := []string{"type", "val"}
	expected := `{"data":{"type":"val"}}`

	// Execute SUT
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonCompliantFormatSimpleKeyWithIdStringValue(t *testing.T) {
	// Fixture Setup
	input := []string{"id", "val"}
	expected := `{"data":{"id":"val"}}`

	// Execute SUT
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonCompliantFormatSimpleNestedKeyValue(t *testing.T) {
	// Fixture Setup
	input := []string{"foo.bar", "val"}
	expected := `{"data":{"attributes":{"foo":{"bar":"val"}}}}`

	// Execute SUT
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonCompliantFormatSimpleKeyNumericValue(t *testing.T) {
	// Fixture Setup
	input := []string{"key", "3"}
	expected := `{"data":{"attributes":{"key":3}}}`

	// Execute SUT
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonCompliantFormatSimpleKeyBooleanTrueValue(t *testing.T) {
	// Fixture Setup
	input := []string{"key", "true"}
	expected := `{"data":{"attributes":{"key":true}}}`

	// Execute SUT
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonCompliantFormatSimpleKeyBooleanFalseValue(t *testing.T) {
	// Fixture Setup
	input := []string{"key", "false"}
	expected := `{"data":{"attributes":{"key":false}}}`

	// Execute SUT
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonCompliantFormatSimpleKeyNullValue(t *testing.T) {
	// Fixture Setup
	input := []string{"key", "null"}
	expected := `{"data":{"attributes":{"key":null}}}`

	// Execute SUT
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonCompliantFormatSimpleKeyEmptyArrayValue(t *testing.T) {
	// Fixture Setup
	input := []string{"key", "[]"}
	expected := `{"data":{"attributes":{"key":[]}}}`

	// Execute SUT
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}

}

func TestToJsonCompliantFormatSimpleArrayIndexValue(t *testing.T) {
	// Fixture Setup
	input := []string{"key[0]", "val"}
	expected := `{"data":{"attributes":{"key":["val"]}}}`

	// Execute SUT
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonCompliantFormatSimpleArrayWithTwoValues(t *testing.T) {
	// Fixture Setup
	input := []string{"key[0]", "val", "key[1]", "val2"}
	expected := `{"data":{"attributes":{"key":["val","val2"]}}}`

	// Execute SUT
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonCompliantForAlias(t *testing.T) {
	// Fixture Setup
	aliases.SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123-456",
		"type": "account"
	}
}`)

	input := []string{"parent_id", "last_read=entity"}

	expected :=
		// language=JSON
		`{"data":{"parent_id":"123-456"}}`

	res, ok := resources.GetResourceByName("accounts")

	require.True(t, ok, "Should have gotten a resource back")

	// Execute SUT
	actual, _ := ToJson(input, false, false, res.Attributes, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonCompliantForAliasWithWildcard(t *testing.T) {
	// Fixture Setup
	aliases.SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123-456",
		"type": "account"
	}
}`)

	input := []string{"data.id", "alias/account/last_read=entity/id"}

	expected :=
		// language=JSON
		`{"data":{"id":"123-456"}}`

	res, ok := resources.GetResourceByName("entries-relationship")

	require.True(t, ok, "Should have gotten a resource back")

	// Execute SUT
	actual, _ := ToJson(input, true, false, res.Attributes, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonCompliantAliasForNestedArray(t *testing.T) {
	// Fixture Setup
	aliases.SaveAliasesForResources(
		// language=JSON
		`
{
	"data": {
		"id": "123-456",
		"type": "product",
		"name": "Clean Shampoo"
	}
}`)
	input := []string{"components.shampoo.options[0].id", "alias/product/name=Clean_Shampoo/id"}
	expected := `{"data":{"attributes":{"components":{"shampoo":{"options":[{"id":"123-456"}]}}}}}`

	// Execute SUT
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonErrorsWhenArrayAndObjectKeysSpecified(t *testing.T) {
	// Fixture Setup
	input := []string{"[0]", "val", "key", "val2"}
	expected := fmt.Errorf("detected both array syntax arguments '[0]' and object syntax arguments 'key'. Only one format can be used")

	// Execute SUT
	_, actual := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual.Error() != expected.Error() {
		t.Fatalf("Testing json conversion of value '%s' did not match\nExpected: %s\nActually: %s", input, expected, actual)
	}
}

func TestToJsonCreatesSimpleSingleElementArrayWhenArrayKeysSpecified(t *testing.T) {
	// Fixture Setup
	input := []string{"[0]", "val"}
	expected := `{"data":["val"]}`
	// Execute SUT

	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match\nExpected: %s\nActually: %s", input, expected, actual)
	}
}

func TestToJsonCreatesSimpleSingleElementArrayWithNoWrappingWhenArrayKeysSpecified(t *testing.T) {
	// Fixture Setup
	input := []string{"[0]", "val"}
	expected := `["val"]`
	// Execute SUT

	actual, _ := ToJson(input, true, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match\nExpected: %s\nActually: %s", input, expected, actual)
	}
}

func TestToJsonCreatesMultipleElementArrayWhenArrayKeysSpecified(t *testing.T) {
	// Fixture Setup
	input := []string{"[0]", "foo", "[1]", "bar"}
	expected := `{"data":["foo","bar"]}`
	// Execute SUT

	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match\nExpected: %s\nActually: %s", input, expected, actual)
	}
}

func TestToJsonCreatesMultipleElementArrayWhenArrayKeysSpecifiedAndSomeMissing(t *testing.T) {
	// Fixture Setup
	input := []string{"[0]", "foo", "[3]", "bar"}
	expected := `{"data":["foo",null,null,"bar"]}`
	// Execute SUT

	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match\nExpected: %s\nActually: %s", input, expected, actual)
	}
}

func TestToJsonCreatesSimpleSingleElementArrayOfObjectWhenArrayKeysSpecified(t *testing.T) {
	// Fixture Setup
	input := []string{"[0].bar", "val"}
	expected := `{"data":[{"bar":"val"}]}`
	// Execute SUT

	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match\nExpected: %s\nActually: %s", input, expected, actual)
	}
}

func TestToJsonCreatesComplexSingleElementArrayOfObjectWhenArrayKeysSpecified(t *testing.T) {
	// Fixture Setup
	input := []string{"[0].bar", "val", "[1].bar", "tree", "[0].foo", "zoo"}
	expected := `{"data":[{"bar":"val","foo":"zoo"},{"bar":"tree"}]}`
	// Execute SUT

	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match\nExpected: %s\nActually: %s", input, expected, actual)
	}
}

func TestToJsonAddsAdjacentConstantValues(t *testing.T) {
	// Fixture Setup
	input := []string{"data[0].id", "foo", "data[1].id", "bar", "data[0].slug", "first", "data[2].slug", "third"}
	expected := `{"data":[{"id":"foo","slug":"first","type": "hello-world"},{"id":"bar","type":"hello-world"},{"slug":"third","type":"hello-world"}]}`
	// Execute SUT

	actual, _ := ToJson(input, true, false, map[string]*resources.CrudEntityAttribute{
		"data[n].type": {
			Key:  "data[n].type",
			Type: "CONST:hello-world",
		},
		"data[n].id": {
			Key:  "data[n].id",
			Type: "STRING",
		},
		"data[n].slug": {
			Key:  "data[n].slug",
			Type: "STRING",
		},
	}, true)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match\nExpected: %s\nActually: %s", input, expected, actual)
	}
}
