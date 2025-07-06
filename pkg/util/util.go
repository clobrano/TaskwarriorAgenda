package util

import (
	"fmt"
	"log"
	"strings"
	"time"

	googleclient "github.com/clobrano/TaskwarriorAgenda/pkg/google"
	"github.com/clobrano/TaskwarriorAgenda/pkg/taskwarrior"
	"google.golang.org/api/calendar/v3"
)

const (
	NEEDS_UPDATE_DESCRIPTION = "description"
	NEEDS_UPDATE_STATUS      = "status"
	NEEDS_UPDATE_DUE         = "due"
)

// getDateRange return the range of date (from date, to date) in a list of Taskwarrior.Tasks
func GetDateRange(tasks []taskwarrior.Task) (fromDate time.Time, toDate time.Time) {
	now := time.Now()
	fromDate = time.Now()
	toDate = time.Now()
	for _, t := range tasks {
		if t.Due == nil || t.Due.IsZero() {
			continue
		}
		if t.Due.Before(now) {
			if fromDate.After(t.Due.Time) {
				// Start 1 day before
				fromDate = t.Due.Time.Add(-24 * time.Hour)
			}
		} else {
			if toDate.Before(t.Due.Time) {
				// End 1 day after
				toDate = t.Due.Time.Add(24 * time.Hour)
			}
		}
	}
	return fromDate, toDate
}

// EventNeedsUpdate returns true if the fields shared between a taskwarrior.Task and a calendar.Event differ
func EventNeedsUpdate(task *taskwarrior.Task, event *calendar.Event) (bool, string, error) {
	task.EventID = event.Id

	// NOTE: calendar.Event's Summary might have a leading "check" sign () if the task was completed
	originalSummary, summaryWithoutCheckmark, found := strings.Cut(event.Summary, googleclient.EVENT_COMPLETED_PREFIX)
	if found {
		originalSummary = strings.TrimSpace(originalSummary)
		summaryWithoutCheckmark = strings.TrimSpace(summaryWithoutCheckmark)

		// events with checkmark in their description are expected to match completed tasks only
		if task.Status == taskwarrior.PENDING {
			return true, NEEDS_UPDATE_DESCRIPTION, nil
		}
		if task.Description != summaryWithoutCheckmark {
			log.Printf("task: '%s', event: '%s'  needs update\n", task.Description, summaryWithoutCheckmark)
			return true, NEEDS_UPDATE_DESCRIPTION, nil
		}
		return false, "", nil
	}

	if task.Description != event.Summary {
		log.Printf("task: '%s', event: '%s'  needs update\n", task.Description, originalSummary)
		return true, NEEDS_UPDATE_DESCRIPTION, nil
	}

	eventTime, err := time.Parse(time.RFC3339, event.Start.DateTime)
	if err != nil {
		return false, "", err
	}

	if eventTime.After(task.Due.Time) || eventTime.Before(task.Due.Time) {
		log.Printf("task: %s:%s:%s, event: %s:%s time needs update\n", task.Description, task.UUID, task.Status, event.Summary, event.Description)
		return true, NEEDS_UPDATE_DUE, nil
	}

	if !strings.Contains(event.Description, fmt.Sprintf("status: %s", task.Status)) {
		log.Printf("task: %s:%s:%s, event: %s:%s status needs update\n", task.Description, task.UUID, task.Status, event.Summary, event.Description)
		return true, NEEDS_UPDATE_STATUS, nil
	}
	return false, "", nil
}

func ConvertTaskwarriorTaskToCalendarEvent(task *taskwarrior.Task) (*calendar.Event, error) {
	if task == nil {
		return nil, fmt.Errorf("could not convert nil Task")
	}

	if task.Due.IsZero() {
		return nil, fmt.Errorf("could not sync Task without due date: task uuid %s\n", task.UUID)
	}

	var eventSummary string
	var eventStatus string

	switch task.Status {
	case taskwarrior.PENDING:
		eventSummary = task.Description
		eventStatus = "confirmed"
	case taskwarrior.COMPLETED:
		eventSummary = fmt.Sprintf("%s %s", googleclient.EVENT_COMPLETED_PREFIX, task.Description)
		eventStatus = "confirmed"
	case taskwarrior.DELETED:
		eventSummary = fmt.Sprintf("%s %s", googleclient.EVENT_DELETED_PREFIX, task.Description)
		eventStatus = "cancelled"
	default:
		eventSummary = task.Description
		eventStatus = "confirmed"
	}

	defaultDuration := 30 * time.Minute
	eventDescription := fmt.Sprintf("Taskwarrior uuid: %s, status: %s", task.UUID, task.Status)

	event := &calendar.Event{
		Id:      task.EventID,
		Summary: eventSummary,
		Status:  eventStatus,
		Start: &calendar.EventDateTime{
			DateTime: task.Due.UTC().Format(time.RFC3339),
			// TimeZone: "UTC", // Or time.Local.String() if you prefer local, but UTC is usually better for sync
		},
		End: &calendar.EventDateTime{
			DateTime: task.Due.UTC().Add(defaultDuration).Format(time.RFC3339),
			// TimeZone: "UTC",
		},
		Description: eventDescription,
	}

	return event, nil
}
