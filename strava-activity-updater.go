//go:build ignore
// +build ignore

package main

//lint:ignore U1000 This is a main program file
import (
	"flag"
	"log"
	"os"

	"strava-activity-updater/auth"
	"strava-activity-updater/strava"
)

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

	// Get latest activity
	activity, err := strava.GetLatestActivity(config.AccessToken)
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
		update := strava.ActivityUpdate{
			Name:      "Pickup Ice Hockey",
			SportType: "IceSkate",
		}

		// Update the activity
		if err := strava.UpdateActivity(config.AccessToken, activity.ID, update); err != nil {
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
