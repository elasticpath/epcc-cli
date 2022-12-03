package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/authentication"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/profiles"
	"github.com/elasticpath/epcc-cli/external/shutdown"
	"github.com/elasticpath/epcc-cli/external/version"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

var RawHeaders []string

const EnvNameHttpPrefix = "EPCC_CLI_HTTP_HEADER_"

var httpHeaders = map[string]string{}

var DontLog2xxs = false

var stats = struct {
	totalRateLimitedTimeInMs       int64
	totalHttpRequestProcessingTime int64
	totalRequests                  uint64

	respCodes map[int]int
}{}

func init() {
	stats.respCodes = make(map[int]int)
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

	go func() {
		lastTotalRequests := uint64(0)

		for {
			time.Sleep(15 * time.Second)

			statsLock.Lock()

			deltaRequests := stats.totalRequests - lastTotalRequests
			lastTotalRequests = stats.totalRequests
			statsLock.Unlock()

			if shutdown.ShutdownFlag.Load() {
				break
			}

			if deltaRequests > 0 {
				log.Infof("Total requests %d, requests in past 15 seconds %d, latest %d requests per second.", lastTotalRequests, deltaRequests, deltaRequests/15.0)
			}

		}
	}()
}

var Limit *rate.Limiter = nil

var Retry429 = false
var Retry5xx = false

var statsLock = &sync.Mutex{}

var HttpClient = &http.Client{}

func LogStats() {
	statsLock.Lock()
	defer statsLock.Unlock()
	keys := make([]int, 0, len(stats.respCodes))

	for k := range stats.respCodes {
		keys = append(keys, k)
	}

	sort.Ints(keys)

	counts := ""

	for _, k := range keys {
		counts += fmt.Sprintf("%d:%d, ", k, stats.respCodes[k])
	}

	if stats.totalRequests > 3 {
		log.Infof("Total requests %d, and total rate limiting time %d ms, and total processing time %d ms. Response Code Count: %s", stats.totalRequests, stats.totalRateLimitedTimeInMs, stats.totalHttpRequestProcessingTime, counts)
	} else {
		log.Debugf("Total requests %d, and total rate limiting time %d ms and total processing time %d ms. Response Code Count: %s", stats.totalRequests, stats.totalRateLimitedTimeInMs, stats.totalHttpRequestProcessingTime, counts)
	}
}
func DoRequest(ctx context.Context, method string, path string, query string, payload io.Reader) (response *http.Response, error error) {
	return doRequestInternal(ctx, method, "application/json", path, query, payload)
}

func DoFileRequest(ctx context.Context, path string, payload io.Reader, contentType string) (response *http.Response, error error) {
	return doRequestInternal(ctx, "POST", contentType, path, "", payload)
}

var UserAgent = fmt.Sprintf("epcc-cli/%s-%s (%s/%s)", version.Version, version.Commit, runtime.GOOS, runtime.GOARCH)

var noApiEndpointUrlWarningMessageMutex = sync.RWMutex{}

var noApiEndpointUrlWarningMessageLogged = false

// DoRequest makes a html request to the EPCC API and handles the response.
func doRequestInternal(ctx context.Context, method string, contentType string, path string, query string, payload io.Reader) (response *http.Response, error error) {

	if shutdown.ShutdownFlag.Load() {
		return nil, fmt.Errorf("Shutting down")
	}

	reqURL, err := url.Parse(config.Envs.EPCC_API_BASE_URL)
	if err != nil {
		return nil, err
	}

	if reqURL.Host == "" {
		noApiEndpointUrlWarningMessageMutex.RLock()
		// Double check lock, read once with read lock, then once again with write lock
		if !noApiEndpointUrlWarningMessageLogged {
			noApiEndpointUrlWarningMessageMutex.RUnlock()
			noApiEndpointUrlWarningMessageMutex.Lock()
			if !noApiEndpointUrlWarningMessageLogged {
				log.Infof("No API endpoint set in profile or environment variables, defaulting to \"%s\". To change this set the EPCC_API_BASE_URL environment variable.", config.DefaultUrl)
				noApiEndpointUrlWarningMessageLogged = true
			}
			noApiEndpointUrlWarningMessageMutex.Unlock()
		} else {
			noApiEndpointUrlWarningMessageMutex.RUnlock()
		}

		reqURL, err = url.Parse(config.DefaultUrl)
		if err != nil {
			log.Fatalf("Error when parsing default host, this is a bug, %s", config.DefaultUrl)
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

	rateLimitTime := time.Since(start)
	resp, err := HttpClient.Do(req)
	requestTime := time.Since(start)

	statsLock.Lock()
	stats.totalRequests += 1
	if rateLimitTime.Milliseconds() > 50 {
		// Only count rate limit time if it took us longer than 50 ms to get here.
		stats.totalRateLimitedTimeInMs += int64(rateLimitTime.Milliseconds())
	}

	stats.totalHttpRequestProcessingTime += int64(requestTime.Milliseconds()) - int64(rateLimitTime.Milliseconds())

	if resp != nil {
		stats.respCodes[resp.StatusCode] = stats.respCodes[resp.StatusCode] + 1
	}

	requestNumber := stats.totalRequests
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

	displayLongFormRequestAndResponse := resp.StatusCode >= 400

	if resp.StatusCode == 429 && Retry429 {
		displayLongFormRequestAndResponse = false
	}

	if resp.StatusCode >= 500 && Retry5xx {
		displayLongFormRequestAndResponse = false
	}

	displayLongFormRequestAndResponse = displayLongFormRequestAndResponse || log.IsLevelEnabled(log.DebugLevel)

	if displayLongFormRequestAndResponse {
		if payload != nil {
			body, _ := io.ReadAll(&bodyBuf)
			if len(body) > 0 {
				logf("(%0.4d) %s %s%s", requestNumber, method, reqURL.String(), requestHeaders)
				if contentType == "application/json" {
					json.PrintJsonToStderr(string(body))
				} else {
					logf("%s", body)
				}

				logf("%s %s%s", resp.Proto, resp.Status, responseHeaders)
			} else {
				logf("(%0.4d) %s %s%s ==> %s %s%s", requestNumber, req.Method, getUrl(reqURL), requestHeaders, resp.Proto, resp.Status, responseHeaders)
			}
		} else {
			logf("(%0.4d) %s %s%s ==> %s %s%s", requestNumber, req.Method, getUrl(reqURL), requestHeaders, resp.Proto, resp.Status, responseHeaders)
		}
	} else {
		if resp.StatusCode >= 300 || !DontLog2xxs {
			log.Infof("(%0.4d) %s %s ==> %s %s", requestNumber, method, getUrl(reqURL), resp.Proto, resp.Status)
		}
	}

	dumpRes, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Error(err)
	}

	profiles.LogRequestToDisk(method, path, dumpReq, dumpRes, resp.StatusCode)

	if resp.StatusCode == 429 && Retry429 {
		return doRequestInternal(ctx, method, contentType, path, query, &bodyBuf)
	} else if resp.StatusCode >= 500 && Retry5xx {
		return doRequestInternal(ctx, method, contentType, path, query, &bodyBuf)
	} else {
		return resp, err
	}

}

func getUrl(u *url.URL) string {
	query, _ := url.PathUnescape(u.String())
	return query
	//query, _ := url.PathUnescape(u.RawQuery)
	//return fmt.Sprintf("%s://%s/%s?%s", u.Scheme, u.Host, u.Path, query)
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
