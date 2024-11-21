package oidc

import (
	"context"
	"encoding/base64"
	gojson "encoding/json"
	"fmt"
	"github.com/elasticpath/epcc-cli/external/authentication"
	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/rest"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
)

type TokenPageInfo struct {
	LoginType        string
	ErrorTitle       string
	ErrorDescription string
	Name             string
	Id               string
}

func GetTokenData(ctx context.Context, port uint16, r *http.Request) (*TokenPageInfo, error) {
	// Parse the query parameters
	queryParams := r.URL.Query()

	// Convert query parameters to a map
	data := make(map[string]string)
	for key, values := range queryParams {
		data[key] = values[0] // Use the first value if multiple are provided
	}

	if data["login_type"] == "AM" {
		token := data["token"]
		amTokenJson, err := base64.URLEncoding.DecodeString(token)

		if err != nil {
			return nil, fmt.Errorf("could not get decode am token: %w", err)
		}

		amToken := authentication.AccountManagementAuthenticationTokenStruct{}

		err = gojson.Unmarshal(amTokenJson, &amToken)

		if err != nil {
			return nil, fmt.Errorf("could not get unmarshal am token: %w", err)
		}

		authentication.SaveAccountManagementAuthenticationToken(amToken)

		apiToken := authentication.GetApiToken()

		if apiToken != nil {
			if apiToken.Identifier == "client_credentials" {
				log.Warnf("You are currently logged in with client_credentials, please switch to implicit with `epcc login implicit` to use the account management token correctly. Mixing client_credentials and the account management token can lead to unintended results.")
			}
		}

		go func() {
			time.Sleep(2 * time.Second)
			log.Infof("Authentication complete, shutting down")
			os.Exit(0)
		}()

		return &TokenPageInfo{
			LoginType: "AM",
			Name:      amToken.AccountName,
			Id:        amToken.AccountId,
		}, nil

	} else if data["login_type"] == "Customers" {
		token := data["token"]
		custToken, err := base64.URLEncoding.DecodeString(token)

		if err != nil {
			return nil, fmt.Errorf("could not get decode am token: %w", err)
		}

		custTokenStruct := authentication.CustomerTokenStruct{}

		err = gojson.Unmarshal(custToken, &custTokenStruct)

		if err != nil {
			return nil, fmt.Errorf("could not get unmarshal am token: %w", err)
		}

		ctr := authentication.CustomerTokenResponse{
			Data:           custTokenStruct,
			AdditionalInfo: authentication.CustomerTokenEpccCliAdditionalInfo{},
		}

		authentication.SaveCustomerToken(ctr)

		apiToken := authentication.GetApiToken()

		if apiToken != nil {
			if apiToken.Identifier == "client_credentials" {
				log.Warnf("You are currently logged in with client_credentials, please switch to implicit with `epcc login implicit` to use the customer token correctly. Mixing client_credentials and the customer token can lead to unintended results.")
			}
		}

		if authentication.IsAccountManagementAuthenticationTokenSet() {
			log.Warnf("Logging out of Account Management")
			authentication.ClearAccountManagementAuthenticationToken()
		}

		result, err := rest.GetInternal(context.Background(), &httpclient.HttpParameterOverrides{}, []string{"customer", custTokenStruct.CustomerId}, false)

		customerName := "Unknown"
		customerEmail := "Unkwown"

		if err == nil {
			customerName, err = json.RunJQOnStringAndGetString(".data.name", result)

			if err != nil {
				log.Warnf("Could not get customer name from response %s, %v", result, err)
			}

			customerEmail, err = json.RunJQOnStringAndGetString(".data.email", result)

			if err != nil {
				log.Warnf("Could not get customer email from response %s, %v", result, err)
			}

			ctr := authentication.CustomerTokenResponse{
				Data: custTokenStruct,
				AdditionalInfo: authentication.CustomerTokenEpccCliAdditionalInfo{
					CustomerName:  customerName,
					CustomerEmail: customerEmail,
				},
			}

			log.Infof("Saving customer token with %s,%s, %v", customerName, customerEmail, result)
			authentication.SaveCustomerToken(ctr)
		}

		go func() {
			time.Sleep(2 * time.Second)
			log.Infof("Authentication complete, shutting down")
			os.Exit(0)
		}()

		return &TokenPageInfo{
			LoginType: "Customers",
			Name:      customerName,
			Id:        custTokenStruct.CustomerId,
		}, nil

	}

	return nil, fmt.Errorf("invalid login type")
}
