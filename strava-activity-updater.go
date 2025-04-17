package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Strict type definitions for all structs
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

type Activity struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	SportType   string    `json:"sport_type"`
	StartDate   time.Time `json:"start_date"`
	Description string    `json:"description"`
}

type ActivityUpdate struct {
	Name        string `json:"name,omitempty"`
	SportType   string `json:"sport_type,omitempty"`
	Description string `json:"description,omitempty"`
}

// Main function with entry point
func main() {
	// Parse command line arguments
	apiKeyPtr := flag.String("api-key", "", "Strava API key")
	configFilePtr := flag.String("config", "strava_config.json", "Path to config file")
	verbosePtr := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	// Set up logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime)

	// Load configuration
	config, err := loadConfig(*configFilePtr)
	if err != nil {
		log.Printf("Could not load config file, will attempt to create it")
		config = &StravaConfig{}
	}

	// Set API key from command line if provided
	if *apiKeyPtr != "" {
		config.RefreshToken = *apiKeyPtr
	}

	if config.RefreshToken == "" {
		log.Fatalf("No refresh token provided. Please specify either via config file or -api-key flag")
	}

	// Ensure we have a valid access token
	if err := ensureValidToken(config); err != nil {
		log.Fatalf("Failed to obtain valid token: %v", err)
	}

	// Save updated config
	if err := saveConfig(*configFilePtr, config); err != nil {
		log.Printf("Warning: Failed to save config: %v", err)
	}

	// Get latest activity
	activity, err := getLatestActivity(config.AccessToken)
	if err != nil {
		log.Fatalf("Failed to get latest activity: %v", err)
	}

	if *verbosePtr {
		log.Printf("Latest activity: ID=%d, Name='%s', Type='%s'",
			activity.ID, activity.Name, activity.SportType)
	}

	// Check if we need to update the activity
	if activity.Name == "Morning Workout" && activity.SportType == "Workout" {
		// Create update payload
		update := ActivityUpdate{
			Name:      "Pickup Ice Hockey",
			SportType: "IceSkate",
		}

		// Update the activity
		if err := updateActivity(config.AccessToken, activity.ID, update); err != nil {
			log.Fatalf("Failed to update activity: %v", err)
		}

		log.Printf("Successfully updated activity ID %d:", activity.ID)
		log.Printf("  - Changed Name from '%s' to '%s'", activity.Name, update.Name)
		log.Printf("  - Changed Sport Type from '%s' to '%s'", activity.SportType, update.SportType)
	} else {
		log.Printf("No update needed for activity ID %d", activity.ID)
		if *verbosePtr {
			log.Printf("  Current Name: '%s'", activity.Name)
			log.Printf("  Current Sport Type: '%s'", activity.SportType)
		}
	}
}

// ensureValidToken checks if the current token is valid,
// if not it refreshes the token
func ensureValidToken(config *StravaConfig) error {
	// Check if token is expired or missing
	if config.AccessToken == "" || time.Now().Unix() >= config.ExpiresAt {
		// Need to refresh the token
		return refreshToken(config)
	}
	return nil
}

// refreshToken gets a new access token using the refresh token
func refreshToken(config *StravaConfig) error {
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

	// Update the config with the new token
	config.AccessToken = tokenResp.AccessToken
	config.RefreshToken = tokenResp.RefreshToken
	config.ExpiresAt = tokenResp.ExpiresAt

	return nil
}

// getLatestActivity fetches the most recent activity for the authenticated user
func getLatestActivity(accessToken string) (*Activity, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://www.strava.com/api/v3/athlete/activities?per_page=1", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get activities: %s - %s", resp.Status, string(body))
	}

	var activities []Activity
	if err := json.NewDecoder(resp.Body).Decode(&activities); err != nil {
		return nil, fmt.Errorf("failed to decode activities: %w", err)
	}

	if len(activities) == 0 {
		return nil, fmt.Errorf("no activities found")
	}

	return &activities[0], nil
}

// updateActivity updates an existing activity with new information
func updateActivity(accessToken string, activityID int64, update ActivityUpdate) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Convert update to JSON
	updateJSON, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal update: %w", err)
	}

	// Create request
	url := fmt.Sprintf("https://www.strava.com/api/v3/activities/%d", activityID)
	req, err := http.NewRequestWithContext(ctx, "PUT", url,
		strings.NewReader(string(updateJSON)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	// Send request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update activity: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update activity: %s - %s", resp.Status, string(body))
	}

	return nil
}

// loadConfig loads the Strava configuration from a file
func loadConfig(filename string) (*StravaConfig, error) {
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

// saveConfig saves the Strava configuration to a file
func saveConfig(filename string, config *StravaConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0600)
}
