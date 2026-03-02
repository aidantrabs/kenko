package kenko

import "time"

type Target struct {
	Name string
	URL  string
}

type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
)

type Result struct {
	Target     string        `json:"target"`
	URL        string        `json:"url"`
	Status     Status        `json:"status"`
	StatusCode int           `json:"status_code"`
	Latency    time.Duration `json:"latency"`
	Error      string        `json:"error,omitempty"`
	CheckedAt  time.Time     `json:"checked_at"`
}
