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

func PrintJsonToStdout(json string) error {
	defer os.Stdout.Sync()
	return printJsonToWriter(json, shouldPrintMonochrome(), os.Stdout)
}

func shouldPrintMonochrome() bool {
	m := MonochromeOutput
	// Adapted from gojq
	if !m && os.Getenv("TERM") == "dumb" {
		m = true
	} else {
		colorCapableTerminal := isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
		if !colorCapableTerminal {
			m = true
		}
	}

	return m
}

func PrintJsonToWriter(json string, w io.Writer) error {
	return printJsonToWriter(json, true, w)
}

func PrintJsonToStderr(json string) error {
	defer os.Stderr.Sync()
	return printJsonToWriter(json, shouldPrintMonochrome(), os.Stderr)
}

func PrettyPrint(in string) string {
	var out bytes.Buffer
	err := gojson.Indent(&out, []byte(in), "", "   ")
	if err != nil {
		return in
	}
	return out.String()
}

func printJsonToWriter(json string, monoOutput bool, w io.Writer) error {

	var v interface{}

	err := gojson.Unmarshal([]byte(json), &v)

	e := NewEncoder(false, 2, monoOutput)

	done := make(chan bool, 1)

	defer close(done)

	if !monoOutput {
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
