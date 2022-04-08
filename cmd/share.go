package cmd

import (
	b64 "encoding/base64"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

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

func logToDisk(requestMethod string, requestPath string, requestBytes []byte, responseBytes []byte, responseCode int) error {
	logDirectory := "profiles"
	os.Mkdir("profiles", os.ModePerm)
	var logNumber = 1
	lastFile := getLastFile(logDirectory)
	if lastFile != nil {
		decodedFileNAme, err := base64DecodeStripped((*lastFile).Name())
		if err != nil {
			return err
		}

		fileNameParts := strings.Split(decodedFileNAme, " ")
		logNumber, _ = strconv.Atoi(fileNameParts[0])
		logNumber++
	}

	filename := base64EncodeStripped(fmt.Sprintf("%d %s %s ==> %d", logNumber, requestMethod, requestPath, responseCode))
	f, err := os.Create(fmt.Sprintf("%s/%s", logDirectory, filename))
	if err != nil {
		return err
	}
	defer f.Close()
	f.Write(requestBytes)
	f.Write([]byte("\n"))
	f.Write(responseBytes)
	return nil
}

func allFileSortedByDate(logDirectory string) []fs.FileInfo {
	files, _ := ioutil.ReadDir(logDirectory)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})

	return files
}

func getLastFile(logDirectory string) *fs.FileInfo {
	all := allFileSortedByDate(logDirectory)
	if len(all) >= 1 {
		return &all[len(all)-1]
	}
	return nil
}

func base64EncodeStripped(s string) string {
	encoded := b64.StdEncoding.EncodeToString([]byte(s))
	return strings.TrimRight(encoded, "=")
}

func base64DecodeStripped(s string) (string, error) {
	if i := len(s) % 4; i != 0 {
		s += strings.Repeat("=", 4-i)
	}
	decoded, err := b64.StdEncoding.DecodeString(s)
	return string(decoded), err
}
