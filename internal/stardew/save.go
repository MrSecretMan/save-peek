package stardew

import "time"

type Save struct {
	Path       string    `json:"-"`
	Folder     string    `json:"folder"`
	ModifiedAt time.Time `json:"modified_at"`
	Size       int64     `json:"-"`
}

type Progress struct {
	PlayerName      string           `json:"player_name"`
	FarmName        string           `json:"farm_name"`
	Money           int64            `json:"money"`
	Season          string           `json:"season"`
	Day             int              `json:"day"`
	Year            int              `json:"year"`
	PlayTime        int64            `json:"play_time_ms"`
	Skills          Skills           `json:"skills"`
	Achievements    int              `json:"achievements"`
	Relationships   []Friend         `json:"relationships,omitempty"`
	Recommendations []Recommendation `json:"recommendations,omitempty"`
	Source          Source           `json:"source"`
}

type Skills struct {
	Farming  int `json:"farming"`
	Fishing  int `json:"fishing"`
	Foraging int `json:"foraging"`
	Mining   int `json:"mining"`
	Combat   int `json:"combat"`
}

type Friend struct {
	Name   string `json:"name"`
	Points int    `json:"points"`
	Hearts int    `json:"hearts"`
}

type Source struct {
	Folder     string    `json:"folder"`
	ModifiedAt time.Time `json:"modified_at"`
}
