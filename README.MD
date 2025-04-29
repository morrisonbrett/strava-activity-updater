# 🏒 Strava Activity Tools

A collection of tools for managing and analyzing Strava activities.

## Tools

### 1. Activity Counter (`strava-activity-counter.go`)

Counts and displays all your activity names, showing:
- Total count for each unique activity name
- Visual indicators for leading/trailing spaces (→ for leading, ← for trailing, · for internal spaces)
- Sorted by frequency (most common first)

```bash
go run strava-activity-counter.go
```

Example output:
```
Activity Name Counts:
--------------------
Pickup·Ice·Hockey                             300
→Pickup·Ice·Hockey                            150
Pickup·Ice·Hockey←                            100
--------------------
Total unique activities: 3
```

### 2. Activity Renamer (`strava-activity-renamer.go`)

Renames activities based on predefined mappings. Currently configured to:
- Fix capitalization: "Pickup ice Hockey" → "Pickup Ice Hockey"
- Standardize training names: "Private Training Workout" → "Private Training Session"
- Standardize workout names: "Workout w/Trainer" → "Private Training Session"
- Fix typos: "Gym Workou" → "Gym Workout"
- Standardize workout names: "Workout" → "Gym Workout"

```bash
# Show what would be changed (dry run)
go run strava-activity-renamer.go

# Apply the changes
go run strava-activity-renamer.go -dry-run=false
```

### 3. Activity Updater (`strava-activity-updater.go`)

Updates the most recent activity if it matches certain criteria. Currently configured to:
- Change "Morning Workout" to "Pickup Ice Hockey"
- Change sport type from "Workout" to "IceSkate"

```bash
# Run with verbose logging
go run strava-activity-updater.go -verbose
```

## Configuration

All tools use the same configuration file (`strava_config.json`). You can specify a different config file using the `-config` flag:

```bash
go run strava-activity-counter.go -config=my_config.json
```

The config file should contain:
```json
{
  "client_id": "your_client_id",
  "client_secret": "your_client_secret",
  "refresh_token": "your_refresh_token"
}
```

You can also provide the refresh token directly via command line:
```bash
go run strava-activity-counter.go -api-key=your_refresh_token
```

## Common Flags

All tools support these common flags:
- `-api-key`: Strava API key (refresh token)
- `-config`: Path to config file (default: "strava_config.json")
- `-verbose`: Enable verbose logging (where applicable)
- `-dry-run`: Show what would be changed without making changes (where applicable)

## Development

The code is organized into packages:
- `auth`: Authentication and token management
- `strava`: Common types and API functions

Each tool is a separate program that can be run independently.

## 🎯 Purpose

This tool was designed to solve a specific problem: automatically renaming and changing the type of Strava activities after they've been recorded. Perfect for when you regularly record activities that need consistent adjustments.

## ✨ Features

- 🔄 **Idempotent** - Only updates activities that match specific criteria
- 🔐 **Secure** - Handles OAuth token management automatically
- ⏱️ **Automation-ready** - Built for Windows Task Scheduler integration
- 🔒 **Private** - Runs locally, no data sent to third parties

## 🛠️ Setup

### Prerequisites

- Go 1.16+ installed
- A Strava account
- Registered Strava API application

### Installation

```bash
# Clone the repository
git clone https://github.com/morrisonbrett/strava-activity-updater.git
cd strava-activity-updater

# Build the executable
go build
```

### Configuration

1. Register a Strava API application at https://www.strava.com/settings/api
2. Get authorization with the following URL (replace YOUR_CLIENT_ID):
   ```
   https://www.strava.com/oauth/authorize?client_id=YOUR_CLIENT_ID&redirect_uri=http://localhost&response_type=code&scope=activity:read_all,activity:write
   ```
3. Exchange the authorization code for tokens:
   ```
   https://www.strava.com/oauth/token?client_id=YOUR_CLIENT_ID&client_secret=YOUR_CLIENT_SECRET&code=AUTHORIZATION_CODE&grant_type=authorization_code
   ```
4. Create a `strava_config.json` file:
   ```json
   {
     "client_id": "YOUR_CLIENT_ID",
     "client_secret": "YOUR_CLIENT_SECRET",
     "refresh_token": "YOUR_REFRESH_TOKEN",
     "access_token": "",
     "expires_at": 0
   }
   ```

## 🚀 Usage

```bash
# Basic usage
./strava-activity-updater

# With verbose logging
./strava-activity-updater -verbose

# Specify config file location
./strava-activity-updater -config=path/to/config.json
```

### Windows Task Scheduler

1. Open Task Scheduler
2. Create a new Basic Task
3. Set trigger to run at your preferred frequency
4. Action: Start a program
5. Browse to your compiled .exe file
6. Add any command-line arguments
7. Complete the wizard

## ⚙️ Customization

To change what activities get updated, modify the condition in `main.go`:

```go
// Example: Update activities with name starting with "Morning" and type "Workout"
if strings.HasPrefix(activity.Name, "Morning") && activity.SportType == "Workout" {
    // Update details here
}
```

## 📄 License

MIT

---

💡 **Tip**: Use with caution and respect [Strava's API usage policy](https://developers.strava.com/docs/).