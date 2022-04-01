package json

import "testing"

func TestToJsonEmptyValue(t *testing.T) {
	// Fixture Setup
	input := []string{}
	expected := `{"data":{}}`

	// Execute SUT
	actual, _ := ToJson(input)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonSimpleKeyStringValue(t *testing.T) {
	// Fixture Setup
	input := []string{"key", "val"}
	expected := `{"data":{"key":"val"}}`

	// Execute SUT
	actual, _ := ToJson(input)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonSimpleNestedKeyValue(t *testing.T) {
	// Fixture Setup
	input := []string{"foo.bar", "val"}
	expected := `{"data":{"foo":{"bar":"val"}}}`

	// Execute SUT
	actual, _ := ToJson(input)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonSimpleKeyNumericValue(t *testing.T) {
	// Fixture Setup
	input := []string{"key", "3"}
	expected := `{"data":{"key":3}}`

	// Execute SUT
	actual, _ := ToJson(input)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonSimpleKeyBooleanTrueValue(t *testing.T) {
	// Fixture Setup
	input := []string{"key", "true"}
	expected := `{"data":{"key":true}}`

	// Execute SUT
	actual, _ := ToJson(input)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonSimpleKeyBooleanFalseValue(t *testing.T) {
	// Fixture Setup
	input := []string{"key", "false"}
	expected := `{"data":{"key":false}}`

	// Execute SUT
	actual, _ := ToJson(input)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonSimpleKeyNullValue(t *testing.T) {
	// Fixture Setup
	input := []string{"key", "null"}
	expected := `{"data":{"key":null}}`

	// Execute SUT
	actual, _ := ToJson(input)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonSimpleKeyEmptyArrayValue(t *testing.T) {
	// Fixture Setup
	input := []string{"key", "[]"}
	expected := `{"data":{"key":[]}}`

	// Execute SUT
	actual, _ := ToJson(input)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}

}

func TestToJsonSimpleArrayIndexValue(t *testing.T) {
	// Fixture Setup
	input := []string{"key[0]", "val"}
	expected := `{"data":{"key":["val"]}}`

	// Execute SUT
	actual, _ := ToJson(input)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}

func TestToJsonSimpleArrayWithTwoValues(t *testing.T) {
	// Fixture Setup
	input := []string{"key[0]", "val", "key[1]", "val2"}
	expected := `{"data":{"key":["val","val2"]}}`

	// Execute SUT
	actual, _ := ToJson(input)

	// Verification
	if actual != expected {
		t.Fatalf("Testing json conversion of empty value %s did not match expected %s, actually: %s", input, expected, actual)
	}
}
