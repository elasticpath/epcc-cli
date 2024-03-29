package profiles

import (
	b64 "encoding/base64"
	"fmt"
	"io/fs"
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

	if n >= len(files) || n < -len(files) {
		return "", fmt.Errorf("there are only %d entries to show, cannot show entry: %d", len(files), n)
	} else if n < 0 {
		n += len(files)
	}

	dir, err := getRequestLogDirectory()

	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(dir + "/" + files[n].Name())

	if err != nil {
		//Maybe a race condition, but maybe not.
		return "", fmt.Errorf("could not read entry %d, file exists(ed) but failed to read", n)
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

	statusCode := fmt.Sprintf("%d", responseCode)

	if responseCode == 0 {
		statusCode = "ERROR"
	}
	return SaveRequest(fmt.Sprintf("%s %s ==> %s", requestMethod, requestPath, statusCode), requestBytes, responseBytes)
}

func SaveRequest(title string, requestBytes []byte, responseBytes []byte) error {
	titleb64 := base64EncodeStripped(title)

	dir, err := getRequestLogDirectory()

	if err != nil {
		return err
	}

	f, err := os.Create(fmt.Sprintf("%s/%d_%s", dir, time.Now().UnixMicro(), titleb64))
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
		return "", fmt.Errorf("could not make directory %s", dir)
	}

	return dir, nil
}

func allFilesSortedByDate() ([]fs.FileInfo, error) {
	dir, err := getRequestLogDirectory()

	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	infos := make([]fs.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].ModTime().Before(infos[j].ModTime())
	})

	return infos, nil
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
