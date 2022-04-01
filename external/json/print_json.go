package json

import (
	gojson "encoding/json"
	"github.com/mattn/go-isatty"
	"os"
)

var MonochromeOutput = false

func PrintJson(json string) error {
	// Adapted from gojq
	if os.Getenv("TERM") == "dumb" {
		MonochromeOutput = true
	} else {
		colorCapableTerminal := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
		if !colorCapableTerminal {
			MonochromeOutput = true
		}
	}

	var v interface{}

	err := gojson.Unmarshal([]byte(json), &v)

	e := NewEncoder(false, 2)

	err = e.Marshal(v, os.Stdout)
	return err
}
