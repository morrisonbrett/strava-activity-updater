package main

//nolint:gochecknoglobals
import (
	"flag"
	"log"
	"os"

	"strava-activity-updater/auth"
	"strava-activity-updater/strava"
)

var nameMappings = map[string]string{
	"Pickup ice Hockey":        "Pickup Ice Hockey",
	"Private Training Workout": "Private Training Session",
	"Workout w/Trainer":        "Private Training Session",
	"Workout":                  "Gym Workout",
	"Gym Workou":               "Gym Workout",
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
	activities, err := strava.GetAllActivities(config.AccessToken)
	if err != nil {
		log.Fatalf("Failed to get activities: %v", err)
	}

	// Find activities that need to be renamed
	var activitiesToUpdate []strava.Activity
	for _, activity := range activities {
		if _, exists := nameMappings[activity.Name]; exists {
			activitiesToUpdate = append(activitiesToUpdate, activity)
		}
	}

	if len(activitiesToUpdate) == 0 {
		log.Printf("No activities found that need to be renamed")
		return
	}

	// Print what would be changed
	log.Printf("Found %d activities that need to be renamed:", len(activitiesToUpdate))
	for _, activity := range activitiesToUpdate {
		newName := nameMappings[activity.Name]
		log.Printf("  ID: %d", activity.ID)
		log.Printf("    From: '%s'", activity.Name)
		log.Printf("    To:   '%s'", newName)
	}

	if *dryRunPtr {
		log.Printf("\nThis was a dry run. To apply changes, run with -dry-run=false")
		return
	}

	// Apply changes
	log.Printf("\nApplying changes...")
	for _, activity := range activitiesToUpdate {
		newName := nameMappings[activity.Name]
		update := strava.ActivityUpdate{
			Name: newName,
		}

		if err := strava.UpdateActivity(config.AccessToken, activity.ID, update); err != nil {
			log.Printf("Failed to update activity ID %d: %v", activity.ID, err)
			continue
		}

		log.Printf("Successfully updated activity ID %d: '%s' -> '%s'",
			activity.ID, activity.Name, newName)
	}
}
