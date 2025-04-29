package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"strava-activity-updater/auth"
)

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

func main() {
	// Parse command line arguments
	apiKeyPtr := flag.String("api-key", "", "Strava API key")
	configFilePtr := flag.String("config", "strava_config.json", "Path to config file")
	dryRunPtr := flag.Bool("dry-run", true, "Show what would be changed without making changes")
	flag.Parse()

	// Set up logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime)

	// Load configuration
	config, err := auth.LoadConfig(*configFilePtr)
	if err != nil {
		log.Printf("Could not load config file, will attempt to create it")
		config = &auth.StravaConfig{}
	}

	// Set API key from command line if provided
	if *apiKeyPtr != "" {
		config.RefreshToken = *apiKeyPtr
	}

	if config.RefreshToken == "" {
		log.Fatalf("No refresh token provided. Please specify either via config file or -api-key flag")
	}

	// Ensure we have a valid access token
	if err := auth.EnsureValidToken(config); err != nil {
		log.Fatalf("Failed to obtain valid token: %v", err)
	}

	// Save updated config
	if err := auth.SaveConfig(*configFilePtr, config); err != nil {
		log.Printf("Warning: Failed to save config: %v", err)
	}

	// Get all activities
	activities, err := getAllActivities(config.AccessToken)
	if err != nil {
		log.Fatalf("Failed to get activities: %v", err)
	}

	// Find activities with leading/trailing spaces
	var activitiesToUpdate []Activity
	for _, activity := range activities {
		trimmedName := strings.TrimSpace(activity.Name)
		if trimmedName != activity.Name {
			activitiesToUpdate = append(activitiesToUpdate, activity)
		}
	}

	if len(activitiesToUpdate) == 0 {
		log.Printf("No activities found with leading or trailing spaces")
		return
	}

	// Print what would be changed
	log.Printf("Found %d activities with leading or trailing spaces:", len(activitiesToUpdate))
	for _, activity := range activitiesToUpdate {
		trimmedName := strings.TrimSpace(activity.Name)
		log.Printf("  ID: %d", activity.ID)
		log.Printf("    From: '%s'", activity.Name)
		log.Printf("    To:   '%s'", trimmedName)
	}

	if *dryRunPtr {
		log.Printf("\nThis was a dry run. To apply changes, run with -dry-run=false")
		return
	}

	// Apply changes
	log.Printf("\nApplying changes...")
	for _, activity := range activitiesToUpdate {
		trimmedName := strings.TrimSpace(activity.Name)
		update := ActivityUpdate{
			Name: trimmedName,
		}

		if err := updateActivity(config.AccessToken, activity.ID, update); err != nil {
			log.Printf("Failed to update activity ID %d: %v", activity.ID, err)
			continue
		}

		log.Printf("Successfully updated activity ID %d: '%s' -> '%s'",
			activity.ID, activity.Name, trimmedName)
	}
}

func getAllActivities(accessToken string) ([]Activity, error) {
	var allActivities []Activity
	page := 1
	perPage := 200 // Maximum allowed by Strava API

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		url := fmt.Sprintf("https://www.strava.com/api/v3/athlete/activities?per_page=%d&page=%d", perPage, page)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Add("Authorization", "Bearer "+accessToken)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to get activities: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("failed to get activities: %s - %s", resp.Status, string(body))
		}

		var activities []Activity
		if err := json.NewDecoder(resp.Body).Decode(&activities); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode activities: %w", err)
		}
		resp.Body.Close()

		if len(activities) == 0 {
			break
		}

		allActivities = append(allActivities, activities...)
		page++

		// If we got fewer activities than requested, we've reached the end
		if len(activities) < perPage {
			break
		}
	}

	return allActivities, nil
}

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
