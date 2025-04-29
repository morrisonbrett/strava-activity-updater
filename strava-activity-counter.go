package main

//lint:ignore U1000 This is a main program file
import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"strava-activity-updater/auth"
	"strava-activity-updater/strava"
)

func main() {
	// Parse command line arguments
	apiKeyPtr := flag.String("api-key", "", "Strava API key")
	configFilePtr := flag.String("config", "strava_config.json", "Path to config file")
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

	// Count activities by name
	activityCounts := make(map[string]int)
	for _, activity := range activities {
		activityCounts[activity.Name]++
	}

	// Convert to slice for sorting
	type ActivityCount struct {
		Name  string
		Count int
	}
	var counts []ActivityCount
	for name, count := range activityCounts {
		counts = append(counts, ActivityCount{name, count})
	}

	// Sort by count (descending)
	sort.Slice(counts, func(i, j int) bool {
		return counts[i].Count > counts[j].Count
	})

	// Print results
	fmt.Printf("\nActivity Name Counts:\n")
	fmt.Printf("--------------------\n")
	for _, count := range counts {
		// Visualize spaces in the name
		visualizedName := strings.ReplaceAll(count.Name, " ", "·")
		if strings.HasPrefix(count.Name, " ") {
			visualizedName = "→" + visualizedName
		}
		if strings.HasSuffix(count.Name, " ") {
			visualizedName = visualizedName + "←"
		}
		fmt.Printf("%-40s %d\n", visualizedName, count.Count)
	}
	fmt.Printf("--------------------\n")
	fmt.Printf("Total unique activities: %d\n", len(counts))
}
