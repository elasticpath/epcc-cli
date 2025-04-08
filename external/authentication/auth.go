package authentication

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/profiles"
	"github.com/elasticpath/epcc-cli/external/version"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type ApiTokenResponse struct {
	Expires     int64  `json:"expires"`
	ExpiresIn   int    `json:"expires_in"`
	Identifier  string `json:"identifier"`
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
}

var HttpClient = &http.Client{
	Timeout: time.Second * 60,
}

var bearerToken atomic.Pointer[ApiTokenResponse]

var noTokenWarningMutex = sync.RWMutex{}

var noTokenWarningMessageLogged = false

var getTokenMutex = sync.Mutex{}

var postAuthErrorHook = []func(r *http.Request, e error){}
var postAuthHook = []func(r *http.Request, s *http.Response){}

func AddPostAuthErrorHook(f func(r *http.Request, e error)) {
	getTokenMutex.Lock()
	defer getTokenMutex.Unlock()
	postAuthErrorHook = append(postAuthErrorHook, f)
}

func AddPostAuthHook(f func(r *http.Request, s *http.Response)) {
	getTokenMutex.Lock()
	defer getTokenMutex.Unlock()
	postAuthHook = append(postAuthHook, f)
}
func GetAuthenticationToken(useTokenFromProfileDir bool, valuesOverride *url.Values, warnOnNoAuthentication bool) (*ApiTokenResponse, error) {

	if useTokenFromProfileDir {
		bearerToken.Store(GetApiToken())
	}

	bearerTokenVal := bearerToken.Load()

	if bearerTokenVal != nil {
		if time.Now().Unix()+60 < bearerTokenVal.Expires {
			// Use cached authentication (but clone first)
			bearerCopy := *bearerTokenVal
			return &bearerCopy, nil
		}
	}

	getTokenMutex.Lock()
	defer getTokenMutex.Unlock()

	if bearerTokenVal != nil {
		if time.Now().Unix()+60 < bearerTokenVal.Expires {
			// Use cached authentication (but clone first)
			bearerCopy := *bearerTokenVal
			return &bearerCopy, nil
		} else {
			// TODO This will also happen a bunch of times in concurrent goroutines
			log.Infof("Existing token has expired (or will very soon), refreshing. Token expiry is at %s", time.Unix(bearerTokenVal.Expires, 0).Format(time.RFC1123Z))
		}
	}

	env := config.GetEnv()
	requestValues := valuesOverride
	if requestValues == nil {
		if IsAutoLoginEnabled() {
			values := url.Values{}
			var grantType string

			// Autologin using env vars
			if env.EPCC_CLIENT_ID == "" {
				noTokenWarningMutex.RLock()
				// Double check lock, read once with read lock, then once again with write lock
				if noTokenWarningMessageLogged == false {
					noTokenWarningMutex.RUnlock()
					noTokenWarningMutex.Lock()
					defer noTokenWarningMutex.Unlock()
					if noTokenWarningMessageLogged == false {
						noTokenWarningMessageLogged = true
						if !env.EPCC_CLI_SUPPRESS_NO_AUTH_MESSAGES && warnOnNoAuthentication {
							log.Warn("No client id set in profile or env var, no authentication will be used for API request. To get started, set the EPCC_CLIENT_ID and (optionally) EPCC_CLIENT_SECRET environment variables")
						}

					}
				} else {
					noTokenWarningMutex.RUnlock()
				}

				return nil, nil
			}

			values.Set("client_id", env.EPCC_CLIENT_ID)
			grantType = "implicit"

			clientSecret := env.EPCC_CLIENT_SECRET
			if clientSecret != "" {
				values.Set("client_secret", clientSecret)
				grantType = "client_credentials"
			}

			values.Set("grant_type", grantType)

			requestValues = &values
		} else {

			noTokenWarningMutex.RLock()
			if noTokenWarningMessageLogged == false {
				// Double check lock, read once with read lock, then once again with write lock
				noTokenWarningMutex.RUnlock()
				noTokenWarningMutex.Lock()
				defer noTokenWarningMutex.Unlock()
				if noTokenWarningMessageLogged == false {
					noTokenWarningMessageLogged = true
					if !config.GetEnv().EPCC_CLI_SUPPRESS_NO_AUTH_MESSAGES {
						log.Infof("Automatic login is disabled, re-enable by using `epcc login client_credentials`")
					}
				}
			} else {
				noTokenWarningMutex.RUnlock()
			}

			return nil, nil
		}
	} else {
		if !IsAutoLoginEnabled() {
			err := EnableAutoLogin()
			if err == nil {
				log.Infof("Re-enabling automatic login")
			} else {
				log.Warnf("Could not enable automatic login %v", err)
			}
		}

	}
	token, err := fetchNewAuthenticationToken(*requestValues)

	if err != nil {
		return nil, err
	}

	bearerToken.Store(token)

	SaveApiToken(token)

	return token, nil
}

// fetchNewAuthenticationToken returns an AccessToken or an Error
func fetchNewAuthenticationToken(values url.Values) (*ApiTokenResponse, error) {

	reqURL, err := url.Parse(config.GetEnv().EPCC_API_BASE_URL)
	if err != nil {
		return nil, err
	}

	if reqURL.Host == "" {
		log.Infof("No API endpoint set in profile or environment variables, defaulting to \"%s\". To change this set the EPCC_API_BASE_URL environment variable. (2)", config.DefaultUrl)
		reqURL, err = url.Parse(config.DefaultUrl)
		if err != nil {
			log.Fatalf("Error when parsing default host, this is a bug, %s", config.DefaultUrl)
		}
	}

	if reqURL.Host == "api.moltin.com" {
		log.Warnf("The API Endpoint https://api.moltin.com is deprecated, please use https://euwest.api.elasticpath.com instead")
	}

	reqURL.Path = fmt.Sprintf("/oauth/access_token")

	body := strings.NewReader(values.Encode())

	req, err := http.NewRequest("POST", reqURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", fmt.Sprintf("epcc-cli/%s-%s", version.Version, version.Commit))

	dumpReq, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Errorf("error %v", err)
	}

	resp, err := HttpClient.Do(req)

	if err != nil {
		for _, f := range postAuthErrorHook {
			f(req, err)
		}

		return nil, err
	}

	defer resp.Body.Close()

	for _, f := range postAuthHook {
		f(req, resp)
	}

	dumpRes, _ := httputil.DumpResponse(resp, true)

	profiles.LogRequestToDisk("POST", req.URL.Path, dumpReq, dumpRes, resp.StatusCode)

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

	var buffer bytes.Buffer
	_, err = buffer.ReadFrom(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 || log.IsLevelEnabled(log.DebugLevel) {
		logf("%s %s%s", "POST", reqURL.String(), requestHeaders)
		logf("%s", values.Encode())
		logf("%s %s%s", resp.Proto, resp.Status, responseHeaders)
		logf("%s", buffer.Bytes())
	} else if resp.StatusCode >= 200 && resp.StatusCode <= 399 {
		log.Infof("%s %s ==> %s %s", "POST", reqURL.String(), resp.Proto, resp.Status)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error: unexpected status %s", resp.Status)
	}

	var authResponse ApiTokenResponse
	if err := json.Unmarshal(buffer.Bytes(), &authResponse); err != nil {
		return nil, err
	}

	log.Trace("Authentication successful")

	return &authResponse, nil
}
