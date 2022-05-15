package json

import (
	"bytes"
	"fmt"
	wincolor "github.com/gookit/color"
	"io"
	"math"
	"math/big"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// This whole file is 98% based on being copied from https://github.com/itchyny/gojq/blob/main/cli/encoder.go
// https://github.com/itchyny/gojq/blob/main/cli/color.go

type encoder struct {
	out      io.Writer
	w        *bytes.Buffer
	tab      bool
	indent   int
	depth    int
	buf      [64]byte
	keyStack []string
}

type colorInfo struct {
	unix        []byte
	colorString string
}

func setColor(buf *bytes.Buffer, color colorInfo) {
	if !MonochromeOutput {
		buf.Write([]byte(color.colorString))
	}
}

func newColor(unix string, colorString string) colorInfo {
	var colorByte = []byte(nil)
	if unix != "" {
		colorByte = []byte("\x1b[" + unix + "m")
	}
	return colorInfo{
		unix:        colorByte,
		colorString: colorString,
	}
}

var unimportantPrefixes = map[string]bool{
	"data":            true,
	"data.attributes": true,
	"data.links":      true,
	"data.meta":       true,
	//"data.meta.created_at":                      true,
	//"data.meta.updated_at":                      true,
	"data.relationships":                        true,
	"data.relationships.children":               true,
	"data.relationships.children.data":          true,
	"data.relationships.children.links":         true,
	"data.relationships.children.links.related": true,
	"links":              true,
	"links.current":      true,
	"links.first":        true,
	"links.last":         true,
	"links.next":         true,
	"links.prev":         true,
	"meta":               true,
	"meta.page":          true,
	"meta.page.current":  true,
	"meta.page.limit":    true,
	"meta.page.offset":   true,
	"meta.page.total":    true,
	"meta.results.total": true,
	"meta.results":       true,
}

var importantPrefixes = map[string]bool{}

var urgentPrefixes = map[string]bool{
	"errors.id":     true,
	"errors.status": true,
	"errors.detail": true,
	"errors.title":  true,
	"errors":        true,
}

var (
	resetColor                = newColor("0", "</>")                     // Reset
	nullColor                 = newColor("90", "<gray>")                 // Bright black
	falseColor                = newColor("33", "<yellow>")               // Yellow
	trueColor                 = newColor("33", "<yellow>")               // Yellow
	numberColor               = newColor("36", "<cyan>")                 // Cyan
	stringColor               = newColor("32", "<green>")                // Green
	objectKeyColor            = newColor("34;1", "<fg=blue;op=bold>")    // Bold Blue
	unimportantObjectKeyColor = newColor("34", "<blue>")                 // Blue
	importantObjectKeyColor   = newColor("35;1", "<fg=magenta;op=bold>") // Bold Purple
	urgentObjectKeyColor      = newColor("31;1", "<fg=red;op=bold>")     // Bold Red
	arrayColor                = newColor("", "<default>")                // No color
	objectColor               = newColor("", "<default>")                // No color
)

func NewEncoder(tab bool, indent int) *encoder {
	// reuse the buffer in multiple calls of marshal
	return &encoder{w: new(bytes.Buffer), tab: tab, indent: indent}
}

func (e *encoder) Marshal(v interface{}, w io.Writer) error {
	e.out = w
	e.encode(v)
	wincolor.Fprint(w, string(e.w.Bytes()))
	e.w.Reset()

	return nil
}

func (e *encoder) encode(v interface{}) {
	switch v := v.(type) {
	case nil:
		e.write([]byte("null"), &nullColor)
	case bool:
		if v {
			e.write([]byte("true"), &trueColor)
		} else {
			e.write([]byte("false"), &falseColor)
		}
	case int:
		e.write(strconv.AppendInt(e.buf[:0], int64(v), 10), &numberColor)
	case float64:
		e.encodeFloat64(v)
	case *big.Int:
		e.write(v.Append(e.buf[:0], 10), &numberColor)
	case string:
		e.encodeString(v, &stringColor)
	case []interface{}:
		e.encodeArray(v)
	case map[string]interface{}:
		e.encodeMap(v)
	default:
		panic(fmt.Sprintf("invalid value: %v", v))
	}
	// Original code to prevent buffering, but if we are outputting color this will break
	/*if e.w.Len() > 8*1024 {
		e.out.Write(e.w.Bytes())
		e.w.Reset()
	}*/
}

// ref: floatEncoder in encoding/json
func (e *encoder) encodeFloat64(f float64) {
	if math.IsNaN(f) {
		e.write([]byte("null"), &nullColor)
		return
	}
	if f >= math.MaxFloat64 {
		f = math.MaxFloat64
	} else if f <= -math.MaxFloat64 {
		f = -math.MaxFloat64
	}
	fmt := byte('f')
	if x := math.Abs(f); x != 0 && x < 1e-6 || x >= 1e21 {
		fmt = 'e'
	}
	buf := strconv.AppendFloat(e.buf[:0], f, fmt, -1, 64)
	if fmt == 'e' {
		// clean up e-09 to e-9
		if n := len(buf); n >= 4 && buf[n-4] == 'e' && buf[n-3] == '-' && buf[n-2] == '0' {
			buf[n-2] = buf[n-1]
			buf = buf[:n-1]
		}
	}
	e.write(buf, &numberColor)
}

// ref: encodeState#string in encoding/json
func (e *encoder) encodeString(s string, color *colorInfo) {
	if color != nil {
		setColor(e.w, *color)
	}
	e.w.WriteByte('"')
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if ' ' <= b && b <= '~' && b != '"' && b != '\\' {
				i++
				continue
			}
			if start < i {
				e.w.WriteString(s[start:i])
			}
			e.w.WriteByte('\\')
			switch b {
			case '\\', '"':
				e.w.WriteByte(b)
			case '\b':
				e.w.WriteByte('b')
			case '\f':
				e.w.WriteByte('f')
			case '\n':
				e.w.WriteByte('n')
			case '\r':
				e.w.WriteByte('r')
			case '\t':
				e.w.WriteByte('t')
			default:
				const hex = "0123456789abcdef"
				e.w.WriteString("u00")
				e.w.WriteByte(hex[b>>4])
				e.w.WriteByte(hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				e.w.WriteString(s[start:i])
			}
			e.w.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		e.w.WriteString(s[start:])
	}
	e.w.WriteByte('"')
	if color != nil {
		setColor(e.w, resetColor)
	}
}

func (e *encoder) encodeArray(vs []interface{}) {
	e.writeByte('[', &arrayColor)
	e.depth += e.indent
	for i, v := range vs {
		if i > 0 {
			e.writeByte(',', &arrayColor)
		}
		if e.indent != 0 {
			e.writeIndent()
		}
		e.encode(v)
	}
	e.depth -= e.indent
	if len(vs) > 0 && e.indent != 0 {
		e.writeIndent()
	}
	e.writeByte(']', &arrayColor)
}

func (e *encoder) encodeMap(vs map[string]interface{}) {
	e.writeByte('{', &objectColor)
	e.depth += e.indent
	type keyVal struct {
		key string
		val interface{}
	}
	kvs := make([]keyVal, len(vs))
	var i int
	for k, v := range vs {
		kvs[i] = keyVal{k, v}
		i++
	}
	sort.Slice(kvs, func(i, j int) bool {

		if kvs[i].key == "type" {
			return true
		} else if kvs[j].key == "type" {
			return false
		}

		if kvs[i].key == "id" {
			return true
		} else if kvs[j].key == "id" {
			return false
		}

		return kvs[i].key < kvs[j].key
	})
	for i, kv := range kvs {
		if i > 0 {
			e.writeByte(',', &objectColor)
		}
		if e.indent != 0 {
			e.writeIndent()
		}

		old := e.keyStack
		e.keyStack = append(e.keyStack, kv.key)

		prefix := strings.Join(e.keyStack, ".")

		keyColorToUse := objectKeyColor

		if _, ok := unimportantPrefixes[prefix]; ok {
			keyColorToUse = unimportantObjectKeyColor
		}

		if _, ok := importantPrefixes[prefix]; ok {
			keyColorToUse = importantObjectKeyColor
		}

		if _, ok := urgentPrefixes[prefix]; ok {
			keyColorToUse = urgentObjectKeyColor
		}

		e.encodeString(kv.key, &keyColorToUse)
		e.writeByte(':', &objectColor)
		if e.indent != 0 {
			e.w.WriteByte(' ')
		}

		e.encode(kv.val)

		e.keyStack = old

	}
	e.depth -= e.indent
	if len(vs) > 0 && e.indent != 0 {
		e.writeIndent()
	}
	e.writeByte('}', &objectColor)
}

func (e *encoder) writeIndent() {
	e.w.WriteByte('\n')
	if n := e.depth; n > 0 {
		if e.tab {
			const tabs = "\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t"
			for n > len(tabs) {
				e.w.Write([]byte(tabs))
				n -= len(tabs)
			}
			e.w.Write([]byte(tabs)[:n])
		} else {
			const spaces = "                                                                "
			for n > len(spaces) {
				e.w.Write([]byte(spaces))
				n -= len(spaces)
			}
			e.w.Write([]byte(spaces)[:n])
		}
	}
}

func (e *encoder) writeByte(b byte, color *colorInfo) {
	if color == nil {
		e.w.WriteByte(b)
	} else {
		setColor(e.w, *color)
		e.w.WriteByte(b)
		setColor(e.w, resetColor)
	}
}

func (e *encoder) write(bs []byte, color *colorInfo) {
	if color == nil {
		e.w.Write(bs)
	} else {
		setColor(e.w, *color)
		e.w.Write(bs)
		setColor(e.w, resetColor)
	}
}
