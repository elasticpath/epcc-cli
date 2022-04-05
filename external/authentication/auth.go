package authentication

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/elasticpath/epcc-cli/config"
	"github.com/elasticpath/epcc-cli/external/version"
	log "github.com/sirupsen/logrus"

	"net/http"
	"net/url"
	"strings"
	"time"
)

type authResponse struct {
	Expires     int    `json:"expires"`
	ExpiresIn   int    `json:"expires_in"`
	Identifier  string `json:"identifier"`
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
}

var HttpClient = &http.Client{
	Timeout: time.Second * 10,
}

var bearerToken = ""
var authTime = time.Now()

func GetAuthenticationToken() (string, error) {

	if bearerToken != "" && time.Now().Sub(authTime).Minutes() < 30 {
		// Use cached authentication
		return bearerToken, nil
	}

	token, err := auth()

	if err != nil {
		return "", err
	}

	bearerToken = token
	authTime = time.Now()

	return bearerToken, nil
}

//auth returns an AccessToken or an Error
func auth() (string, error) {
	reqURL, err := url.Parse(config.Envs.EPCC_API_BASE_URL)

	reqURL.Path = fmt.Sprintf("/oauth/access_token")

	values := url.Values{}
	values.Set("client_id", config.Envs.EPCC_CLIENT_ID)
	grantType := "implicit"

	if config.Envs.EPCC_CLIENT_SECRET != "" {
		values.Set("client_secret", config.Envs.EPCC_CLIENT_SECRET)
		grantType = "client_credentials"
	}

	values.Set("grant_type", grantType)

	body := strings.NewReader(values.Encode())

	req, err := http.NewRequest("POST", reqURL.String(), body)
	if err != nil {
		return "", err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", fmt.Sprintf("epcc-cli/%s-%s", version.Version, version.Commit))

	resp, err := HttpClient.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("error: unexpected status %s", resp.Status)
	}

	var buffer bytes.Buffer
	_, err = buffer.ReadFrom(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var authResponse authResponse
	if err := json.Unmarshal(buffer.Bytes(), &authResponse); err != nil {
		return "", err
	}

	log.Trace("Authentication successful")
	return authResponse.AccessToken, nil
}