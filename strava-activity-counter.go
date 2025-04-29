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

	// Count activities by name and sport type
	activityCounts := make(map[string]int)
	sportTypeCounts := make(map[string]int)
	for _, activity := range activities {
		activityCounts[activity.Name]++
		sportTypeCounts[activity.SportType]++
	}

	// Convert to slices for sorting
	type Count struct {
		Name  string
		Count int
	}
	var nameCounts []Count
	var sportTypeCountsList []Count

	for name, count := range activityCounts {
		nameCounts = append(nameCounts, Count{name, count})
	}
	for sportType, count := range sportTypeCounts {
		sportTypeCountsList = append(sportTypeCountsList, Count{sportType, count})
	}

	// Sort by count (descending)
	sort.Slice(nameCounts, func(i, j int) bool {
		return nameCounts[i].Count > nameCounts[j].Count
	})
	sort.Slice(sportTypeCountsList, func(i, j int) bool {
		return sportTypeCountsList[i].Count > sportTypeCountsList[j].Count
	})

	// Print name counts
	fmt.Printf("\nActivity Name Counts:\n")
	fmt.Printf("--------------------\n")
	for _, count := range nameCounts {
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
	fmt.Printf("Total unique activities: %d\n", len(nameCounts))

	// Print sport type counts
	fmt.Printf("\nSport Type Counts:\n")
	fmt.Printf("--------------------\n")
	for _, count := range sportTypeCountsList {
		fmt.Printf("%-40s %d\n", count.Name, count.Count)
	}
	fmt.Printf("--------------------\n")
	fmt.Printf("Total unique sport types: %d\n", len(sportTypeCountsList))
}
