package strava

import "time"

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
