package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/authentication"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/profiles"
	"github.com/elasticpath/epcc-cli/external/version"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var RawHeaders []string

const EnvNameHttpPrefix = "EPCC_CLI_HTTP_HEADER_"

var httpHeaders = map[string]string{}

func init() {
	for _, env := range os.Environ() {
		splitEnv := strings.SplitN(env, "=", 2)

		if len(splitEnv) == 2 {
			envName := splitEnv[0]
			envValue := splitEnv[1]
			if strings.HasPrefix(envName, EnvNameHttpPrefix) {
				headersSplit := strings.SplitN(envValue, ":", 2)

				if len(headersSplit) != 2 {
					log.Warnf("Found environment variable with malformed value %s => %s. Headers should be set in a Key: Value format. This value is being ignored.", envName, envValue)
				} else {
					httpHeaders[headersSplit[0]] = headersSplit[1]
				}
			}
		}
	}
}

var Limit *rate.Limiter = nil

var statsLock = &sync.Mutex{}

const defaultUrl = "https://api.moltin.com"

var stats = struct {
	totalRateLimitedTimeInMs int64
	totalRequests            uint64
}{}

var HttpClient = &http.Client{}

func LogStats() {
	statsLock.Lock()
	defer statsLock.Unlock()
	if stats.totalRequests > 3 {
		log.Infof("Total requests %d, and total rate limiting time %d ms", stats.totalRequests, stats.totalRateLimitedTimeInMs)
	} else {
		log.Debugf("Total requests %d, and total rate limiting time %d ms", stats.totalRequests, stats.totalRateLimitedTimeInMs)
	}
}
func DoRequest(ctx context.Context, method string, path string, query string, payload io.Reader) (response *http.Response, error error) {
	return doRequestInternal(ctx, method, "application/json", path, query, payload)
}

func DoFileRequest(ctx context.Context, path string, payload io.Reader, contentType string) (response *http.Response, error error) {
	return doRequestInternal(ctx, "POST", contentType, path, "", payload)
}

var UserAgent = fmt.Sprintf("epcc-cli/%s-%s (%s/%s)", version.Version, version.Commit, runtime.GOOS, runtime.GOARCH)

// DoRequest makes a html request to the EPCC API and handles the response.
func doRequestInternal(ctx context.Context, method string, contentType string, path string, query string, payload io.Reader) (response *http.Response, error error) {
	reqURL, err := url.Parse(config.Envs.EPCC_API_BASE_URL)
	if err != nil {
		return nil, err
	}

	if reqURL.Host == "" {
		log.Infof("No API endpoint set in profile or environment variables, defaulting to \"%s\". To change this set the EPCC_API_BASE_URL environment variable.", defaultUrl)
		reqURL, err = url.Parse(defaultUrl)
		if err != nil {
			log.Fatalf("Error when parsing default host, this is a bug, %s", defaultUrl)
		}
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

	bearerToken, err := authentication.GetAuthenticationToken(true, nil)

	if err != nil {
		return nil, err
	}

	if bearerToken != nil {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", bearerToken.AccessToken))
	}

	customerToken := authentication.GetCustomerToken()

	if customerToken != nil {
		req.Header.Add("X-Moltin-Customer-Token", customerToken.Data.Token)
	}

	req.Header.Add("Content-Type", contentType)

	req.Header.Add("User-Agent", UserAgent)

	if len(config.Envs.EPCC_BETA_API_FEATURES) > 0 {
		req.Header.Add("EP-Beta-Features", config.Envs.EPCC_BETA_API_FEATURES)
	}

	if err = AddAdditionalHeadersSpecifiedByFlag(req); err != nil {
		return nil, err
	}

	dumpReq, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Error(err)
	}

	start := time.Now()

	if err := Limit.Wait(ctx); err != nil {
		return nil, fmt.Errorf("Rate limiter returned error %v, %w", err, err)
	}

	elapsed := time.Since(start)
	resp, err := HttpClient.Do(req)

	statsLock.Lock()
	stats.totalRequests += 1
	stats.totalRateLimitedTimeInMs += int64(elapsed.Milliseconds())
	statsLock.Unlock()

	if err != nil {
		return nil, err
	}

	var logf func(string, ...interface{})

	if resp.StatusCode >= 400 {
		logf = func(a string, b ...interface{}) {
			log.Warnf(a, b...)
		}
	} else if log.IsLevelEnabled(log.DebugLevel) {
		logf = func(a string, b ...interface{}) {
			log.Debugf(a, b...)
		}
	} else {
		logf = func(a string, b ...interface{}) {
			// Do nothing
		}
	}

	requestHeaders := ""
	responseHeaders := ""
	if log.IsLevelEnabled(log.DebugLevel) {
		for k, v := range req.Header {
			requestHeaders += "\n" + k + ":" + strings.Join(v, ", ")
		}

		requestHeaders += "\n"
	}

	if log.IsLevelEnabled(log.DebugLevel) {
		for k, v := range resp.Header {
			responseHeaders += "\n" + k + ":" + strings.Join(v, ", ")
		}
		requestHeaders += "\n\n"
	}

	if resp.StatusCode >= 400 || log.IsLevelEnabled(log.DebugLevel) {
		if payload != nil {
			body, _ := ioutil.ReadAll(&bodyBuf)
			if len(body) > 0 {
				logf("%s %s%s", method, reqURL.String(), requestHeaders)
				if contentType == "application/json" {
					json.PrintJsonToStderr(string(body))
				} else {
					logf("%s", body)
				}

				logf("%s %s%s", resp.Proto, resp.Status, responseHeaders)
			} else {
				logf("%s %s%s ==> %s %s%s", req.Method, reqURL.String(), requestHeaders, resp.Proto, resp.Status, responseHeaders)
			}
		} else {
			logf("%s %s%s ==> %s %s%s", req.Method, reqURL.String(), requestHeaders, resp.Proto, resp.Status, responseHeaders)
		}
	} else if resp.StatusCode >= 200 && resp.StatusCode <= 399 {
		log.Infof("%s %s ==> %s %s", method, reqURL.String(), resp.Proto, resp.Status)
	}

	dumpRes, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Error(err)
	}

	profiles.LogRequestToDisk(method, path, dumpReq, dumpRes, resp.StatusCode)

	return resp, err
}

func AddAdditionalHeadersSpecifiedByFlag(r *http.Request) error {
	for _, header := range RawHeaders {
		// Validation and formatting logic for headers could be improved
		entries := strings.Split(header, ":")
		if len(entries) < 2 {
			return fmt.Errorf("header has invalid format")
		}
		r.Header.Set(entries[0], entries[1])
	}

	for key, val := range httpHeaders {
		r.Header.Set(key, val)
	}

	return nil
}
