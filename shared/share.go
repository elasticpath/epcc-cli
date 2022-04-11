package shared

import (
	b64 "encoding/base64"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os/exec"
	"runtime"
	"sort"
	"strings"
)

const LogDirectory = "profiles"

func OpenUrl(cmUrl string) error {
	switch runtime.GOOS {
	case "linux":
		exec.Command("xdg-open", cmUrl).Start()
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", cmUrl).Start()
	case "darwin":
		exec.Command("open", cmUrl).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}

	return nil
}

func AllFilesSortedByDate(logDirectory string) []fs.FileInfo {
	files, _ := ioutil.ReadDir(logDirectory)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})

	return files
}

func Base64EncodeStripped(s string) string {
	encoded := b64.StdEncoding.EncodeToString([]byte(s))
	return strings.TrimRight(encoded, "=")
}

func Base64DecodeStripped(s string) (string, error) {
	if i := len(s) % 4; i != 0 {
		s += strings.Repeat("=", 4-i)
	}
	decoded, err := b64.StdEncoding.DecodeString(s)
	return string(decoded), err
}
