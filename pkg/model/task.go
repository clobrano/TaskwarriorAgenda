package model

import "time"

// Task represents a generic task from any source.
type Task struct {
	ID          string
	Description string
	Deadline    time.Time
	Tags        []string
	Priority    string
	Status      string
	Source      string // "taskwarrior" or "orgmode"
}
