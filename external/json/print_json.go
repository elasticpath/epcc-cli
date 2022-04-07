package json

import (
	gojson "encoding/json"
	"github.com/mattn/go-isatty"
	"io"
	"os"
)

var MonochromeOutput = false

func PrintJson(json string) error {
	defer os.Stdout.Sync()
	return printJsonToWriter(json, os.Stdout)

}

func PrintJsonToStderr(json string) error {
	defer os.Stderr.Sync()
	return printJsonToWriter(json, os.Stderr)
}

func printJsonToWriter(json string, w io.Writer) error {
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

	err = e.Marshal(v, w)
	return err
}
