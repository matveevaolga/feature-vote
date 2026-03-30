package dto

import "time"

type HealthResponse struct {
	Status    string       `json:"status"`
	Timestamp time.Time    `json:"timestamp"`
	Checks    HealthChecks `json:"checks"`
}

type HealthChecks struct {
	Database string `json:"database"`
	Votings  string `json:"votings"`
	Uptime   string `json:"uptime"`
}

type HealthErrorResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error"`
}
