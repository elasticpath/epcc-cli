package oidc

import (
	"context"
	"encoding/base64"
	gojson "encoding/json"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/authentication"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/rest"
	"net/http"
)

type CallbackPageInfo struct {
	LoginType                 string
	ErrorTitle                string
	ErrorDescription          string
	AccountTokenResponse      *authentication.AccountManagementAuthenticationTokenResponse
	CustomerTokenResponse     *authentication.CustomerTokenResponse
	AccountTokenStructBase64  []string
	CustomerTokenStructBase64 string
}

func GetCallbackData(ctx context.Context, port uint16, r *http.Request) (*CallbackPageInfo, error) {
	// Parse the query parameters
	queryParams := r.URL.Query()

	// Convert query parameters to a map
	data := make(map[string]string)
	for key, values := range queryParams {
		data[key] = values[0] // Use the first value if multiple are provided
	}

	data["uri"] = fmt.Sprintf("http://localhost:%d", port)

	if data["code"] != "" {

		state, err := r.Cookie("state")

		if err != nil {
			return nil, fmt.Errorf("could not get state cookie: %w", err)
		}

		verifier, err := r.Cookie("code_verifier")

		if err != nil {
			return nil, fmt.Errorf("could not get verifier cookie: %w", err)
		}

		login_type, err := r.Cookie("login_type")

		if err != nil {
			return nil, fmt.Errorf("could not get login_type cookie: %w", err)
		}

		if data["state"] != state.Value {
			return &CallbackPageInfo{
				ErrorTitle:       "State Mismatch",
				ErrorDescription: "State mismatch between locally stored value and value from IdP",
				LoginType:        login_type.Value,
			}, nil
		}

		cpi := CallbackPageInfo{
			LoginType: login_type.Value,
		}

		if login_type.Value == "AM" {
			result, err := rest.CreateInternal(context.Background(), &httpclient.HttpParameterOverrides{}, []string{"account-management-authentication-token",
				"authentication_mechanism", "oidc",
				"oauth_authorization_code", data["code"],
				"oauth_redirect_uri", fmt.Sprintf("http://localhost:%d/callback", port),
				"oauth_code_verifier", verifier.Value,
			}, false, "", true, false)

			if err != nil {
				return nil, fmt.Errorf("could not get account tokens: %w", err)
			}

			err = gojson.Unmarshal([]byte(result), &cpi.AccountTokenResponse)

			if err != nil {
				return nil, fmt.Errorf("could not unmarshal response: %w", err)
			}

			for _, v := range cpi.AccountTokenResponse.Data {

				str, err := gojson.Marshal(v)

				if err != nil {
					return nil, fmt.Errorf("could not encode token: %w", err)
				}

				cpi.AccountTokenStructBase64 = append(cpi.AccountTokenStructBase64, base64.URLEncoding.EncodeToString(str))
			}

			return &cpi, nil
		} else if login_type.Value == "Customers" {
			result, err := rest.CreateInternal(context.Background(), &httpclient.HttpParameterOverrides{}, []string{"customer-token",
				"authentication_mechanism", "oidc",
				"oauth_authorization_code", data["code"],
				"oauth_redirect_uri", fmt.Sprintf("http://localhost:%d/callback", port),
				"oauth_code_verifier", verifier.Value,
			}, false, "", true, false)

			if err != nil {
				return nil, fmt.Errorf("could not get customer tokens: %w", err)
			}

			err = gojson.Unmarshal([]byte(result), &cpi.CustomerTokenResponse)

			str, err := gojson.Marshal(cpi.CustomerTokenResponse.Data)

			if err != nil {
				return nil, fmt.Errorf("could not encode token: %w", err)
			}

			cpi.CustomerTokenStructBase64 = base64.URLEncoding.EncodeToString(str)
			return &cpi, nil
		} else {
			return &CallbackPageInfo{
				ErrorTitle:       "Unknown Login Type",
				ErrorDescription: fmt.Sprintf("Unsupported login type used: %v", login_type.Value),
			}, nil
		}
	} else if data["error"] == "" {

		return &CallbackPageInfo{
			ErrorTitle:       "Bad Response",
			ErrorDescription: "Invalid response from IdP, no code or error query parameter",
		}, nil
	} else {
		return &CallbackPageInfo{
			ErrorTitle:       data["error"],
			ErrorDescription: data["error_description"],
		}, nil
	}
}
