package variables

import (
	"sync"
	"testing"
)

func setupTest(t *testing.T) {
	t.Helper()
	InitializeDirectoryForTesting()
	t.Cleanup(func() {
		ClearTestState()
	})
}

func TestSetAndGetVariable(t *testing.T) {
	setupTest(t)

	SetVariable("foo", "bar")
	val, ok := GetVariable("foo")
	if !ok {
		t.Fatal("expected variable 'foo' to exist")
	}
	if val != "bar" {
		t.Fatalf("expected 'bar', got %q", val)
	}
}

func TestGetVariableNotFound(t *testing.T) {
	setupTest(t)

	val, ok := GetVariable("nonexistent")
	if ok {
		t.Fatal("expected variable to not exist")
	}
	if val != "" {
		t.Fatalf("expected empty string, got %q", val)
	}
}

func TestGetAllVariables(t *testing.T) {
	setupTest(t)

	SetVariable("a", "1")
	SetVariable("b", "2")

	all := GetAllVariables()
	if len(all) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(all))
	}
	if all["a"] != "1" || all["b"] != "2" {
		t.Fatalf("unexpected values: %v", all)
	}

	// Ensure it's a copy
	all["c"] = "3"
	if _, ok := GetVariable("c"); ok {
		t.Fatal("modifying returned map should not affect internal state")
	}
}

func TestResolveVariableOrReturnIdentity_WithPrefix(t *testing.T) {
	setupTest(t)

	SetVariable("myid", "12345")
	result := ResolveVariableOrReturnIdentity("var/myid")
	if result != "12345" {
		t.Fatalf("expected '12345', got %q", result)
	}
}

func TestResolveVariableOrReturnIdentity_WithoutPrefix(t *testing.T) {
	setupTest(t)

	result := ResolveVariableOrReturnIdentity("some_plain_value")
	if result != "some_plain_value" {
		t.Fatalf("expected 'some_plain_value', got %q", result)
	}
}

func TestResolveVariableOrReturnIdentity_NotFound(t *testing.T) {
	setupTest(t)

	result := ResolveVariableOrReturnIdentity("var/missing")
	if result != "var/missing" {
		t.Fatalf("expected 'var/missing' to be returned as-is, got %q", result)
	}
}

func TestExtractAndSetVariables_String(t *testing.T) {
	setupTest(t)

	body := `{"data":{"id":"abc-123","attributes":{"name":"Test"}}}`
	ExtractAndSetVariables([]string{"myid=.data.id", "myname=.data.attributes.name"}, body)

	val, ok := GetVariable("myid")
	if !ok || val != "abc-123" {
		t.Fatalf("expected 'abc-123', got %q (ok=%v)", val, ok)
	}

	val, ok = GetVariable("myname")
	if !ok || val != "Test" {
		t.Fatalf("expected 'Test', got %q (ok=%v)", val, ok)
	}
}

func TestExtractAndSetVariables_Number(t *testing.T) {
	setupTest(t)

	body := `{"meta":{"results":{"total":42}}}`
	ExtractAndSetVariables([]string{"total=.meta.results.total"}, body)

	val, ok := GetVariable("total")
	if !ok || val != "42" {
		t.Fatalf("expected '42', got %q (ok=%v)", val, ok)
	}
}

func TestExtractAndSetVariables_Boolean(t *testing.T) {
	setupTest(t)

	body := `{"data":{"enabled":true}}`
	ExtractAndSetVariables([]string{"enabled=.data.enabled"}, body)

	val, ok := GetVariable("enabled")
	if !ok || val != "true" {
		t.Fatalf("expected 'true', got %q (ok=%v)", val, ok)
	}
}

func TestExtractAndSetVariables_Null(t *testing.T) {
	setupTest(t)

	body := `{"data":{"value":null}}`
	ExtractAndSetVariables([]string{"val=.data.value"}, body)

	val, ok := GetVariable("val")
	if !ok || val != "" {
		t.Fatalf("expected empty string for null, got %q (ok=%v)", val, ok)
	}
}

func TestExtractAndSetVariables_Object(t *testing.T) {
	setupTest(t)

	body := `{"data":{"attrs":{"a":1,"b":"two"}}}`
	ExtractAndSetVariables([]string{"obj=.data.attrs"}, body)

	val, ok := GetVariable("obj")
	if !ok {
		t.Fatal("expected variable to exist")
	}
	// Should be marshalled JSON
	if val != `{"a":1,"b":"two"}` {
		t.Fatalf("expected JSON object string, got %q", val)
	}
}

func TestExtractAndSetVariables_EmptyResult(t *testing.T) {
	setupTest(t)

	body := `{"data":{}}`
	ExtractAndSetVariables([]string{"val=.data.nonexistent"}, body)

	_, ok := GetVariable("val")
	// null is a valid jq result for missing field, so it stores empty string
	if !ok {
		t.Fatal("expected variable to exist (null -> empty string)")
	}
}

func TestExtractAndSetVariables_InvalidSpec(t *testing.T) {
	setupTest(t)

	body := `{"data":{"id":"123"}}`
	// No "=" in spec — should log warning and skip
	ExtractAndSetVariables([]string{"invalidspec"}, body)

	_, ok := GetVariable("invalidspec")
	if ok {
		t.Fatal("expected variable to not exist for invalid spec")
	}
}

func TestFlushAndReload(t *testing.T) {
	setupTest(t)

	SetVariable("persist", "value123")
	FlushVariables()

	// Reset in-memory state but keep the same directory
	dir := directoryOverride
	mu.Lock()
	vars = map[string]string{}
	loaded.Store(false)
	mu.Unlock()
	directoryOverride = dir

	val, ok := GetVariable("persist")
	if !ok || val != "value123" {
		t.Fatalf("expected 'value123' after reload, got %q (ok=%v)", val, ok)
	}
}

func TestClearAllVariables(t *testing.T) {
	setupTest(t)

	SetVariable("a", "1")
	SetVariable("b", "2")
	ClearAllVariables()

	all := GetAllVariables()
	if len(all) != 0 {
		t.Fatalf("expected 0 variables after clear, got %d", len(all))
	}
}

func TestConcurrentAccess(t *testing.T) {
	setupTest(t)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		name := "var" + string(rune('A'+i%26))
		go func() {
			defer wg.Done()
			SetVariable(name, "value")
		}()
		go func() {
			defer wg.Done()
			GetVariable(name)
		}()
	}
	wg.Wait()
}
