package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/authentication"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/version"
	"github.com/elasticpath/epcc-cli/globals"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var HttpClient = &http.Client{
	Timeout: time.Second * 10,
}

func DoRequest(ctx context.Context, method string, path string, query string, payload io.Reader) (response *http.Response, error error) {
	return doRequestInternal(ctx, method, "application/json", path, query, payload)
}

func DoFileRequest(ctx context.Context, path string, payload io.Reader, contentType string) (response *http.Response, error error) {
	return doRequestInternal(ctx, "POST", contentType, path, "", payload)
}

// DoRequest makes a html request to the EPCC API and handles the response.
func doRequestInternal(ctx context.Context, method string, contentType string, path string, query string, payload io.Reader) (response *http.Response, error error) {
	reqURL, err := url.Parse(config.Envs.EPCC_API_BASE_URL)
	if err != nil {
		return nil, err
	}

	reqURL.Path = path
	reqURL.RawQuery = query

	var bodyBuf bytes.Buffer
	if payload != nil {
		payload = io.TeeReader(payload, &bodyBuf)
	}

	req, err := http.NewRequest(method, reqURL.String(), payload)
	if err != nil {
		return nil, err
	}

	bearerToken, err := authentication.GetAuthenticationToken()

	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", bearerToken))

	req.Header.Add("Content-Type", contentType)

	req.Header.Add("User-Agent", fmt.Sprintf("epcc-cli/%s-%s", version.Version, version.Commit))

	if err = AddHeaderByFlag(req); err != nil {
		return nil, err
	}

	if len(config.Envs.EPCC_BETA_API_FEATURES) > 0 {
		req.Header.Add("EP-Beta-Features", config.Envs.EPCC_BETA_API_FEATURES)
	}

	resp, err := HttpClient.Do(req)

	if resp.StatusCode >= 400 {
		if payload != nil {
			body, _ := ioutil.ReadAll(&bodyBuf)
			if len(body) > 0 {
				log.Warnf("%s %s", method, reqURL.String())

				// TODO maybe check if it's json and if not do something else.
				json.PrintJsonToStderr(string(body))
				log.Warnf("%s %s", resp.Proto, resp.Status)
			} else {
				log.Warnf("%s %s ==> %s %s", method, reqURL.String(), resp.Proto, resp.Status)
			}
		} else {
			log.Warnf("%s %s ==> %s %s", method, reqURL.String(), resp.Proto, resp.Status)
		}

	} else if resp.StatusCode >= 200 && resp.StatusCode <= 399 {
		log.Infof("%s %s ==> %s %s", method, reqURL.String(), resp.Proto, resp.Status)
	}

	return resp, err
}

// https://stackoverflow.com/questions/20205796/post-data-using-the-content-type-multipart-form-data
func EncodeForm(values map[string]string, filename string, paramName string, fileContents []byte) (byteBuf *bytes.Buffer, contentType string, err error) {

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for key, val := range values {
		_ = writer.WriteField(key, val)
	}

	if len(paramName) > 0 {
		part, err := writer.CreateFormFile(paramName, filename)

		if err != nil {
			return nil, "", err
		}

		part.Write(fileContents)
	}

	err = writer.Close()
	if err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}

func AddHeaderByFlag(r *http.Request) error {
	for _, header := range globals.RawHeaders {
		// Validation and formatting logic for headers could be improved
		entries := strings.Split(header, ":")
		if len(entries) < 2 {
			return fmt.Errorf("header has invalid format")
		}
		r.Header.Add(entries[0], entries[1])
	}
	return nil
}
