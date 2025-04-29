package strava

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func GetAllActivities(accessToken string) ([]Activity, error) {
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

func GetLatestActivity(accessToken string) (*Activity, error) {
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

func UpdateActivity(accessToken string, activityID int64, update ActivityUpdate) error {
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
