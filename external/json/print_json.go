package json

import (
	"bytes"
	gojson "encoding/json"
	"github.com/mattn/go-isatty"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
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

func PrettyPrint(in string) string {
	var out bytes.Buffer
	err := gojson.Indent(&out, []byte(in), "", "   ")
	if err != nil {
		return in
	}
	return out.String()
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

	done := make(chan bool, 1)

	if !MonochromeOutput {
		go func() {
			select {
			case <-done:
				break
			case <-time.After(5 * time.Second):
				log.Warnf("Output of JSON has taken more than 5 seconds, you may want to use -M to supress coloring of output ")
			}
		}()
	}

	err = e.Marshal(v, w)
	done <- true

	w.Write([]byte{byte('\n')})
	return err
}
