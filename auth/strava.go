package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

type StravaConfig struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type TokenResponse struct {
	TokenType    string `json:"token_type"`
	AccessToken  string `json:"access_token"`
	ExpiresAt    int64  `json:"expires_at"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

func EnsureValidToken(config *StravaConfig) error {
	if config.AccessToken == "" || time.Now().Unix() >= config.ExpiresAt {
		return RefreshToken(config)
	}
	return nil
}

func RefreshToken(config *StravaConfig) error {
	if config.ClientID == "" || config.ClientSecret == "" {
		return fmt.Errorf("client ID and client secret must be set in the config file")
	}

	data := url.Values{}
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)
	data.Set("refresh_token", config.RefreshToken)
	data.Set("grant_type", "refresh_token")

	resp, err := http.PostForm("https://www.strava.com/oauth/token", data)
	if err != nil {
		return fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to refresh token: %s - %s", resp.Status, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token response: %w", err)
	}

	config.AccessToken = tokenResp.AccessToken
	config.RefreshToken = tokenResp.RefreshToken
	config.ExpiresAt = tokenResp.ExpiresAt

	return nil
}

func LoadConfig(filename string) (*StravaConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config StravaConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func SaveConfig(filename string, config *StravaConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0600)
}
