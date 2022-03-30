package main

import "testing"

func TestNoDuplicateCommands(t *testing.T) {
	set := make(map[string]bool)

	for _, command := range commands {
		set[command.Keyword] = true
	}

	if len(set) != len(commands) {
		t.Fatalf("Duplicate commands have been registered since the length of the keyword set is not the same as the array")
	}
}
