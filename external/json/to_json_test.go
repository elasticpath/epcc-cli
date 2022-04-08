package json

import (
	"github.com/elasticpath/epcc-cli/external/resources"
	"testing"
)

func TestToJsonLegacyFormatEmptyValue(t *testing.T) {
	// Fixture Setup
	input := []string{}
	expected := `{"data":{}}`

	// Execute SUT
	actual, _ := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{})

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
	actual, _ := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{})

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonLegacyFormatSimpleNestedKeyValue(t *testing.T) {
	// Fixture Setup
	input := []string{"foo.bar", "val"}
	expected := `{"data":{"foo":{"bar":"val"}}}`

	// Execute SUT
	actual, _ := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{})

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
	actual, _ := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{})

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
	actual, _ := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{})

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
	actual, _ := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{})

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
	actual, _ := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{})

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
	actual, _ := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{})

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
	actual, _ := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{})

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
	actual, _ := ToJson(input, false, false, map[string]*resources.CrudEntityAttribute{})

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
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{})

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
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{})

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
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{})

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
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{})

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
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{})

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
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{})

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
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{})

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
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{})

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
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{})

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
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{})

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
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{})

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
	actual, _ := ToJson(input, false, true, map[string]*resources.CrudEntityAttribute{})

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}
