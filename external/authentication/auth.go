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
	Timeout: time.Second * 10,
}

var bearerToken *ApiTokenResponse = nil

var noTokenWarningMutex = sync.RWMutex{}

var noTokenWarningMessageLogged = false

func GetAuthenticationToken(useTokenFromProfileDir bool, valuesOverride *url.Values) (*ApiTokenResponse, error) {

	if useTokenFromProfileDir {
		bearerToken = GetApiToken()
	}

	if bearerToken != nil {
		if time.Now().Unix()+60 < bearerToken.Expires {
			// Use cached authentication (but clone first)
			bearerCopy := *bearerToken
			return &bearerCopy, nil
		} else {
			// TODO This will also happen a bunch of times in concurrent goroutines
			log.Infof("Existing token has expired (or will very soon), refreshing. Token expiry is at %s", time.Unix(bearerToken.Expires, 0).Format(time.RFC1123Z))
		}

	}

	requestValues := valuesOverride
	if requestValues == nil {
		if IsAutoLoginEnabled() {
			values := url.Values{}
			var grantType string

			// Autologin using env vars
			if config.Envs.EPCC_CLIENT_ID == "" {
				noTokenWarningMutex.RLock()
				if noTokenWarningMessageLogged == false {
					noTokenWarningMutex.RUnlock()
					noTokenWarningMutex.Lock()
					defer noTokenWarningMutex.Unlock()
					if noTokenWarningMessageLogged == false {
						noTokenWarningMessageLogged = true
						if !config.Envs.EPCC_CLI_SUPPRESS_NO_AUTH_MESSAGES {
							log.Warn("No client id set in profile or env var, no authentication will be used for API request. To get started, set the EPCC_CLIENT_ID and (optionally) EPCC_CLIENT_SECRET environment variables")
						}

					}
				} else {
					noTokenWarningMutex.RUnlock()
				}

				return nil, nil
			}

			values.Set("client_id", config.Envs.EPCC_CLIENT_ID)
			grantType = "implicit"

			if config.Envs.EPCC_CLIENT_SECRET != "" {
				values.Set("client_secret", config.Envs.EPCC_CLIENT_SECRET)
				grantType = "client_credentials"
			}

			values.Set("grant_type", grantType)

			requestValues = &values
		} else {

			noTokenWarningMutex.RLock()
			if noTokenWarningMessageLogged == false {
				noTokenWarningMutex.RUnlock()
				noTokenWarningMutex.Lock()
				defer noTokenWarningMutex.Unlock()
				if noTokenWarningMessageLogged == false {
					noTokenWarningMessageLogged = true
					if !config.Envs.EPCC_CLI_SUPPRESS_NO_AUTH_MESSAGES {
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

	bearerToken = token

	SaveApiToken(bearerToken)
	return bearerToken, nil
}

// fetchNewAuthenticationToken returns an AccessToken or an Error
func fetchNewAuthenticationToken(values url.Values) (*ApiTokenResponse, error) {

	reqURL, err := url.Parse(config.Envs.EPCC_API_BASE_URL)
	if err != nil {
		return nil, err
	}

	if reqURL.Host == "" {
		log.Infof("No API endpoint set in profile or environment variables, defaulting to \"%s\". To change this set the EPCC_API_BASE_URL environment variable.", config.DefaultUrl)
		reqURL, err = url.Parse(config.DefaultUrl)
		if err != nil {
			log.Fatalf("Error when parsing default host, this is a bug, %s", config.DefaultUrl)
		}
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
		return nil, err
	}

	defer resp.Body.Close()

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
