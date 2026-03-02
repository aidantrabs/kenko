package kenko

import "time"

// Target represents an endpoint to be health-checked.
type Target struct {
	Name string
	URL  string
}

// Status represents the outcome of a health check.
type Status string

// Possible Status values.
const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
)

// Result holds the outcome of a single health check against a target.
type Result struct {
	Target     string        `json:"target"`
	URL        string        `json:"url"`
	Status     Status        `json:"status"`
	StatusCode int           `json:"status_code"`
	Latency    time.Duration `json:"latency"`
	Error      string        `json:"error,omitempty"`
	CheckedAt  time.Time     `json:"checked_at"`
}
