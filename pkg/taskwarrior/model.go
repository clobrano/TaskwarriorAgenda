package taskwarrior

import (
	"fmt"
	"strings"
	"time"
)

const (
	PENDING   = "pending"
	COMPLETED = "completed"
	WAITING   = "waiting"
	DELETED   = "deleted"
)

type CustomTime struct {
	time.Time
}

const taskwarriorTimeLayout = "20060102T150405Z" // YYYYMMDDTHHMMSSZ, 'Z' indicates UTC

// UnmarshalJSON implements the json.Unmarshaler interface for CustomTime.
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`) // Remove surrounding quotes
	if s == "" || s == "0" {          // Handle empty string or "0" if Taskwarrior ever outputs it
		ct.Time = time.Time{} // Set to zero value
		return nil
	}

	t, err := time.Parse(taskwarriorTimeLayout, s)
	if err != nil {
		return fmt.Errorf("failed to parse Taskwarrior time string '%s': %w", s, err)
	}
	ct.Time = t
	return nil
}

// MarshalJSON implements the json.Marshaler interface for CustomTime.
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	if ct.Time.IsZero() {
		return []byte(`""`), nil // Export zero time as empty string
	}
	return []byte(`"` + ct.Time.Format(taskwarriorTimeLayout) + `"`), nil
}

type Task struct {
	UUID        string      `json:"uuid"`
	Description string      `json:"description"`
	Due         *CustomTime `json:"due"`
	Scheduled   *CustomTime `json:"scheduled"`
	Status      string      `json:"status"`
	// Only to update corresponding calendar event
	EventID string `json:"event_id"`
}
