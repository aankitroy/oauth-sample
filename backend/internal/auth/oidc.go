// internal/auth/oidc.go
package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope,omitempty"`
}

type OIDCConfig struct {
	TokenURL    string
	ClientID    string
	RedirectURI string
	UserInfoURL string
	// Possibly ClientSecret if needed
}

func ExchangeCodeForTokens(cfg *OIDCConfig, code, codeVerifier string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", cfg.RedirectURI)
	data.Set("client_id", cfg.ClientID)
	data.Set("code_verifier", codeVerifier)
	fmt.Println("data: ", data)
	req, err := http.NewRequest(http.MethodPost, cfg.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	fmt.Println("Request: ", req)

	// If client_secret is needed:
	// req.SetBasicAuth(cfg.ClientID, cfg.ClientSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	fmt.Println("Response: ", resp)
	fmt.Println("Error: ", err)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		return nil, fmt.Errorf("token endpoint error [%d]: %s", resp.StatusCode, buf.String())
	}

	var tr TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, err
	}
	return &tr, nil
}
