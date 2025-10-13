package models

import "time"

type MissionStatus string

const (
	StatusQueued     MissionStatus = "QUEUED"
	StatusInProgress MissionStatus = "IN_PROGRESS"
	StatusCompleted  MissionStatus = "COMPLETED"
	StatusFailed     MissionStatus = "FAILED"
)

type Mission struct {
	ID        string        `json:"mission_id"`
	Payload   string        `json:"payload"`
	Status    MissionStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}
