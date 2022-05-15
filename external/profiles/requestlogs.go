package profiles

import (
	b64 "encoding/base64"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

var SanitizeLogs = true

func GetAllRequestLogTitles() ([]string, error) {
	titles := make([]string, 0)

	files, err := allFilesSortedByDate()

	if err != nil {
		return titles, err
	}
	for i := 0; i < len(files); i++ {
		fname := strings.Split(files[i].Name(), "_")

		if len(fname) >= 2 {
			name, _ := base64DecodeStripped(fname[1])
			titles = append(titles, files[i].ModTime().Format(time.Kitchen)+" "+name)

		} else {
			titles = append(titles, files[i].Name())
		}

	}

	return titles, nil
}

func GetNthRequestLog(n int) (string, error) {

	files, err := allFilesSortedByDate()

	if err != nil {
		return "", err
	}

	if n < 0 {
		return "", fmt.Errorf("You must specify a positive integer log message to show")
	} else if n >= len(files) {
		return "", fmt.Errorf("There are only %d entries to show, cannot show entry: %d", len(files), n)
	}

	dir, err := getRequestLogDirectory()

	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(dir + "/" + files[n].Name())

	if err != nil {
		//Maybe a race condition, but maybe not.
		return "", fmt.Errorf("Could not read entry %d, file exists(ed) but failed to read", n)
	}

	return string(content), nil

}

func ClearAllRequestLogs() error {
	dir, err := getRequestLogDirectory()

	if err != nil {
		return err
	}

	return os.RemoveAll(dir)
}

func LogRequestToDisk(requestMethod string, requestPath string, requestBytes []byte, responseBytes []byte, responseCode int) error {

	if SanitizeLogs {
		regex1 := regexp.MustCompile(`(?i)client_secret\s*[^A-Za-z0-9]\s*[A-Za-z0-9]*`)
		requestBytes = regex1.ReplaceAll(requestBytes, []byte("client_secret=*****"))
		responseBytes = regex1.ReplaceAll(responseBytes, []byte("client_secret=*****"))
	}

	return SaveRequest(fmt.Sprintf("%s %s ==> %d", requestMethod, requestPath, responseCode), requestBytes, responseBytes)
}

func SaveRequest(title string, requestBytes []byte, responseBytes []byte) error {
	titleb64 := base64EncodeStripped(title)

	dir, err := getRequestLogDirectory()

	if err != nil {
		return err
	}

	f, err := os.Create(fmt.Sprintf("%s/%d_%s", dir, time.Now().Unix(), titleb64))
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(requestBytes)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte("\n"))
	if err != nil {
		return err
	}
	_, err = f.Write(responseBytes)
	if err != nil {
		return err
	}

	return nil
}

func getRequestLogDirectory() (string, error) {
	dir := GetProfileDataDirectory() + "/logs"
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("Could not make directory %s", dir)
	}

	return dir, nil
}

func allFilesSortedByDate() ([]fs.FileInfo, error) {
	dir, err := getRequestLogDirectory()

	if err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(dir)

	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().Before(files[j].ModTime())
	})

	return files, nil
}

func base64EncodeStripped(s string) string {
	encoded := b64.URLEncoding.EncodeToString([]byte(s))
	return strings.TrimRight(encoded, "=")
}

func base64DecodeStripped(s string) (string, error) {
	if i := len(s) % 4; i != 0 {
		s += strings.Repeat("=", 4-i)
	}
	decoded, err := b64.URLEncoding.DecodeString(s)
	return string(decoded), err
}
