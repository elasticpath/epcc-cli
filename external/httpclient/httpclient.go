package httpclient

import (
	"bytes"
	"context"
	"fmt"
	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/authentication"
	"github.com/elasticpath/epcc-cli/external/headergroups"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/profiles"
	"github.com/elasticpath/epcc-cli/external/shutdown"
	"github.com/elasticpath/epcc-cli/external/version"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

var RawHeaders []string

const EnvNameHttpPrefix = "EPCC_CLI_HTTP_HEADER_"

const EnvUrlMatch = "EPCC_CLI_URL_MATCH_REGEXP_(\\d+)"

const EnvUrlMatchPrefix = "EPCC_CLI_URL_MATCH_SUBSTITUTION_"

var urlSubstitions = map[*regexp.Regexp]string{}

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

	urlMatchRegexp := regexp.MustCompile(EnvUrlMatch)

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

			if groups := urlMatchRegexp.FindStringSubmatch(envName); groups != nil {
				if groups != nil {
					r, err := regexp.Compile(envValue)

					if err != nil {
						log.Warnf("Environment variable %s has a malformed regex and substition cannot be performed, %v", env, err)
					} else {
						urlSubstitions[r] = os.Getenv(EnvUrlMatchPrefix + groups[1])
					}
				}
			}
		}
	}
}

var Limit *rate.Limiter = nil

func Initialize(rateLimit uint16, requestTimeout float32, statisticsFrequency int) {
	Limit = rate.NewLimiter(rate.Limit(rateLimit), 1)
	HttpClient.Timeout = time.Duration(int64(requestTimeout*1000) * int64(time.Millisecond))

	if statisticsFrequency > 0 {
		go func() {
			lastTotalRequests := uint64(0)

			for {
				time.Sleep(time.Duration(statisticsFrequency) * time.Second)

				statsLock.Lock()

				deltaRequests := stats.totalRequests - lastTotalRequests
				lastTotalRequests = stats.totalRequests
				statsLock.Unlock()

				if shutdown.ShutdownFlag.Load() {
					break
				}

				if deltaRequests > 0 {
					log.Infof("Total requests %d, requests in past %d seconds %d, latest %d requests per second.", lastTotalRequests, statisticsFrequency, deltaRequests, deltaRequests/uint64(statisticsFrequency))
				}

			}
		}()
	}
}

var Retry429 = false
var Retry5xx = false

var RetryConnectionErrors = false

var RetryAllErrors = false
var RetryDelay uint = 500

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
		if k == 0 {
			counts += fmt.Sprintf("CONN_ERROR:%d, ", stats.respCodes[k])
		} else {
			counts += fmt.Sprintf("%d:%d, ", k, stats.respCodes[k])
		}
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

	env := config.GetEnv()

	reqURL, err := url.Parse(env.EPCC_API_BASE_URL)
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
				log.Infof("No API endpoint set in profile or environment variables, defaulting to \"%s\". To change this set the EPCC_API_BASE_URL environment variable (1).", config.DefaultUrl)
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

	origPath := path

	for r, substitution := range urlSubstitions {
		if r.MatchString(path) {
			path = r.ReplaceAllString(path, substitution)
		}
	}

	if origPath != path {
		log.Tracef("URL Replacement transformed %s to %s", origPath, path)
	}

	reqURL.Path = path

	reqURL.RawQuery = query

	var bodyBuf []byte
	if payload != nil {
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(payload)
		if err != nil {
			log.Warnf("Error reading payload, %s", err)
		}
		bodyBuf = buf.Bytes()

		payload = bytes.NewReader(bodyBuf)
	}

	req, err := http.NewRequest(method, reqURL.String(), payload)
	if err != nil {
		return nil, err
	}

	warnOnNoAuthentication := len(headergroups.GetAllHeaderGroups()) == 0

	bearerToken, err := authentication.GetAuthenticationToken(true, nil, warnOnNoAuthentication)

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

	accountManagementAuthenticationToken := authentication.GetAccountManagementAuthenticationToken()

	if accountManagementAuthenticationToken != nil {
		req.Header.Add("EP-Account-Management-Authentication-Token", accountManagementAuthenticationToken.Token)
	}

	req.Header.Add("Content-Type", contentType)

	req.Header.Add("User-Agent", UserAgent)

	if len(env.EPCC_BETA_API_FEATURES) > 0 {
		req.Header.Add("EP-Beta-Features", env.EPCC_BETA_API_FEATURES)
	}

	if err = AddAdditionalHeadersSpecifiedByFlag(req); err != nil {
		return nil, err
	}

	for k, v := range headergroups.GetAllHeaders() {
		req.Header.Add(k, v)
	}

	dumpReq, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Error(err)
	}

	start := time.Now()

	log.Tracef("Waiting for rate limiter")
	if err := Limit.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter returned error %v, %w", err, err)
	}

	rateLimitTime := time.Since(start)
	log.Tracef("Rate limiter allowed call")

	corrID, _ := uuid.NewUUID()

	log.Tracef("Starting HTTP Request %s %s (Correlation ID: %s [not request id])", req.Method, req.URL.String(), corrID.String())
	resp, err := HttpClient.Do(req)
	requestTime := time.Since(start)

	log.Tracef("HTTP Request complete %s %s (Correlation ID: %s [not request id]), duration %d ms", req.Method, req.URL.String(), corrID.String(), requestTime.Milliseconds())
	statsLock.Lock()

	// Lock is not deferred (for perf reasons), so don't
	// forget to unlock it, if you return before it is so.
	stats.totalRequests += 1
	if rateLimitTime.Milliseconds() > 50 {
		// Only count rate limit time if it took us longer than 50 ms to get here.
		stats.totalRateLimitedTimeInMs += rateLimitTime.Milliseconds()
	}

	stats.totalHttpRequestProcessingTime += requestTime.Milliseconds() - rateLimitTime.Milliseconds()

	if resp != nil {
		stats.respCodes[resp.StatusCode] = stats.respCodes[resp.StatusCode] + 1
	} else {
		stats.respCodes[0] = stats.respCodes[0] + 1
	}

	requestNumber := stats.totalRequests
	statsLock.Unlock()

	log.Tracef("Stats processing complete")
	requestError := err
	if requestError != nil {

		resp = &http.Response{
			Status:           "CONNECTION_ERROR",
			StatusCode:       0,
			Proto:            "",
			ProtoMajor:       0,
			ProtoMinor:       0,
			Header:           map[string][]string{},
			Body:             io.NopCloser(strings.NewReader(fmt.Sprintf("%v", err))),
			ContentLength:    0,
			TransferEncoding: nil,
			Close:            true,
			Uncompressed:     false,
			Trailer:          nil,
			Request:          nil,
			TLS:              nil,
		}
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
			body := bodyBuf
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

	log.Tracef("Starting log to disk")
	profiles.LogRequestToDisk(method, path, dumpReq, dumpRes, resp.StatusCode)
	log.Tracef("Done log to disk")
	if resp.StatusCode == 429 && (Retry429 || RetryAllErrors) {
		if RetryDelay > 0 {
			log.Debugf("Retrying request in %d ms", RetryDelay)
			time.Sleep(time.Duration(RetryDelay) * time.Millisecond)
		}
		return doRequestInternal(ctx, method, contentType, path, query, bytes.NewReader(bodyBuf))
	} else if resp.StatusCode >= 500 && (Retry5xx || RetryAllErrors) {
		if RetryDelay > 0 {
			log.Debugf("Retrying request in %d ms", RetryDelay)
			time.Sleep(time.Duration(RetryDelay) * time.Millisecond)
		}
		return doRequestInternal(ctx, method, contentType, path, query, bytes.NewReader(bodyBuf))
	} else if requestError != nil && (RetryConnectionErrors || RetryAllErrors) {
		if RetryDelay > 0 {
			log.Debugf("Retrying request in %d ms", RetryDelay)
			time.Sleep(time.Duration(RetryDelay) * time.Millisecond)
		}
		return doRequestInternal(ctx, method, contentType, path, query, bytes.NewReader(bodyBuf))
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
