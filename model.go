package main

import "time"

type ShortURL struct {
	URL          string     `json:"url,omitempty"`
	ShortURL     string     `json:"short_url,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	LastAccessed *time.Time `json:"last_accessed,omitempty"`
	UniqueVisits int        `json:"unique_visits,omitempty"`
}
