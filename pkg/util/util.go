package util

import (
	"fmt"
	"strings"
	"time"

	googleclient "github.com/clobrano/TaskwarriorAgenda/pkg/google"
	"github.com/clobrano/TaskwarriorAgenda/pkg/taskwarrior"
	"google.golang.org/api/calendar/v3"
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
				// Add 1 day buffer
				fromDate = t.Due.Time.Add(-24 * time.Hour)
			}
		} else {
			if toDate.Before(t.Due.Time) {
				// Add 1 day buffer
				toDate = t.Due.Time.Add(24 * time.Hour)
			}
		}
	}
	return fromDate, toDate
}

// EventNeedsUpdate returns true if the fields shared between a taskwarrior.Task and a calendar.Event differ
func EventNeedsUpdate(task *taskwarrior.Task, event *calendar.Event) (bool, error) {
	task.EventID = event.Id

	// NOTE: calendar.Event's Summary might have a leading "check" sign () if the task was completed
	originalDescription, summaryWithoutCheckmark, found := strings.Cut(event.Summary, googleclient.EVENT_COMPLETED_PREFIX)
	if found {
		// events with checkmark in their description are expected to match completed tasks only
		if task.Status == taskwarrior.PENDING {
			return true, nil
		}
		if task.Description != summaryWithoutCheckmark {
			return true, nil
		}
	}

	if task.Description != originalDescription {
		return true, nil
	}

	eventTime, err := time.Parse(time.RFC3339, event.Start.DateTime)
	if err != nil {
		return false, err
	}

	if eventTime.After(task.Due.Time) || eventTime.Before(task.Due.Time) {
		return true, nil
	}
	return false, nil
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
	eventDescription := fmt.Sprintf("Taskwarrior uuid: %s, status: %s", task.UUID, eventStatus)

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
