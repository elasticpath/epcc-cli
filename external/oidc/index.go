package oidc

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/elasticpath/epcc-cli/external/httpclient"
	"github.com/elasticpath/epcc-cli/external/json"
	"github.com/elasticpath/epcc-cli/external/rest"
	"github.com/google/uuid"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
)

type LoginPageInfo struct {
	CustomerProfiles     []OidcProfileInfo
	CustomerClientId     string
	AccountProfiles      []OidcProfileInfo
	AccountClientId      string
	RedirectUriUnencoded string
	RedirectUriEncoded   string
	State                string
	CodeVerifier         string
	CodeChallenge        string
}

func GetIndexData(ctx context.Context, port uint16) (*LoginPageInfo, error) {

	overrides := &httpclient.HttpParameterOverrides{
		QueryParameters: nil,
		OverrideUrlPath: "",
	}

	// Get customer and account authentication settings to populate the aliases
	customerAuthSettings, err := rest.GetInternal(ctx, overrides, []string{"customer-authentication-settings"}, false, false)

	if err != nil {
		return nil, fmt.Errorf("could not retrieve customer authentication settings: %w", err)
	}

	accountAuthSettings, err := rest.GetInternal(ctx, overrides, []string{"account-authentication-settings"}, false, false)

	if err != nil {
		return nil, fmt.Errorf("could not retrieve account authentication settings: %w", err)
	}

	customerRealmId, err := json.RunJQOnStringAndGetString(".data.relationships[\"authentication-realm\"].data.id", customerAuthSettings)

	if err != nil {
		return nil, fmt.Errorf("could not determine customer realm id: %w", err)
	}

	customerClientId, err := json.RunJQOnStringAndGetString(".data.meta.client_id", customerAuthSettings)

	if err != nil {
		return nil, fmt.Errorf("could not determine customer client id: %w", err)
	}

	accountRealmId, err := json.RunJQOnStringAndGetString(".data.relationships.authentication_realm.data.id", accountAuthSettings)

	if err != nil {
		return nil, fmt.Errorf("could not determine account realm id: %w", err)
	}

	accountClientId, err := json.RunJQOnStringAndGetString(".data.meta.client_id", accountAuthSettings)

	if err != nil {
		return nil, fmt.Errorf("could not determine account client id: %w", err)
	}

	customerProfiles, err := getOidcProfilesForRealm(ctx, overrides, customerRealmId)

	if err != nil {
		return nil, fmt.Errorf("could not get oidc profiles for customers: %w", err)
	}

	accountProfiles, err := getOidcProfilesForRealm(ctx, overrides, accountRealmId)

	if err != nil {
		return nil, fmt.Errorf("could not get oidc profiles for customers: %w", err)
	}

	verifier, err := GenerateCodeVerifier()

	if err != nil {
		return nil, fmt.Errorf("could not get code verifier: %w", err)
	}

	challenge, err := GenerateCodeChallenge(verifier)

	if err != nil {
		return nil, fmt.Errorf("could not get code challenge: %w", err)
	}

	profiles := LoginPageInfo{
		CustomerProfiles:     customerProfiles,
		CustomerClientId:     customerClientId,
		AccountProfiles:      accountProfiles,
		AccountClientId:      accountClientId,
		State:                uuid.New().String(),
		RedirectUriEncoded:   fmt.Sprintf("%s%d%s", "http%3A%2F%2Flocalhost%3A", port, "/callback"),
		RedirectUriUnencoded: fmt.Sprintf("%s%d%s", "http://localhost:", port, "/callback"),
		CodeVerifier:         verifier,
		CodeChallenge:        challenge,
	}

	return &profiles, nil
}

func getOidcProfilesForRealm(ctx context.Context, overrides *httpclient.HttpParameterOverrides, realmId string) ([]OidcProfileInfo, error) {
	res, err := rest.GetInternal(ctx, overrides, []string{"oidc-profiles", realmId}, false, false)

	if err != nil {
		return nil, err
	}

	resObj, err := json.RunJQOnString(".data | map({name: .name, authorization_link: .links[\"authorization-endpoint\"], idp: .meta.issuer})", res)

	if err != nil {
		log.Errorf("Couldn't get oidc profile information: %v", err)
		return nil, err
	}

	var result []OidcProfileInfo

	err = mapstructure.Decode(resObj, &result)

	if err != nil {
		log.Errorf("Couldn't convert into map: %v", err)
		return nil, err
	}

	return result, nil

}

// GenerateCodeVerifier generates a securely random code verifier.
func GenerateCodeVerifier() (string, error) {
	// PKCE requires a code verifier between 43 and 128 characters.
	// We'll generate 32 bytes of randomness, which Base64 encodes to 43 characters.
	verifier := make([]byte, 32)
	_, err := rand.Read(verifier)
	if err != nil {
		return "", fmt.Errorf("failed to generate random verifier: %w", err)
	}

	// Base64 URL encode the random bytes
	return base64.RawURLEncoding.EncodeToString(verifier), nil
}

// GenerateCodeChallenge generates the S256 code challenge from the verifier.
func GenerateCodeChallenge(verifier string) (string, error) {
	// Compute the SHA256 hash of the verifier
	hash := sha256.Sum256([]byte(verifier))

	// Base64 URL encode the hash
	return base64.RawURLEncoding.EncodeToString(hash[:]), nil
}
