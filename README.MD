# 🏒 Strava Activity Updater

A Golang CLI tool that automatically updates your Strava activities to match your preferences.

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