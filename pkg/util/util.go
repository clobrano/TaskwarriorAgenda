package util

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/clobrano/TaskwarriorAgenda/pkg/model"
	"google.golang.org/api/calendar/v3"
)

const (
	NEEDS_UPDATE_DESCRIPTION = "description"
	NEEDS_UPDATE_STATUS      = "status"
	NEEDS_UPDATE_DUE         = "due"
)

// EventNeedsUpdate returns true if the fields shared between a model.Task and a calendar.Event differ
func EventNeedsUpdate(task *model.Task, event *calendar.Event) (bool, string, error) {
	var eventIsCompleted bool
	var eventIsDeleted bool
	var cleanSummary string

	if strings.HasPrefix(event.Summary, "✅") {
		eventIsCompleted = true
		cleanSummary = strings.TrimSpace(strings.TrimPrefix(event.Summary, "✅"))
	} else if strings.HasPrefix(event.Summary, "❌") {
		eventIsDeleted = true
		cleanSummary = strings.TrimSpace(strings.TrimPrefix(event.Summary, "❌"))
	} else {
		cleanSummary = event.Summary
	}

	// Check for status mismatches
	if task.Status == "completed" && !eventIsCompleted {
		return true, NEEDS_UPDATE_STATUS, nil
	}
	if task.Status == "deleted" && !eventIsDeleted {
		return true, NEEDS_UPDATE_STATUS, nil
	}
	if task.Status == "pending" && (eventIsCompleted || eventIsDeleted) {
		return true, NEEDS_UPDATE_STATUS, nil
	}

	// Check for description mismatch
	if task.Description != cleanSummary {
		log.Printf("task: '%s', event: '%s' needs update\n", task.Description, cleanSummary)
		return true, NEEDS_UPDATE_DESCRIPTION, nil
	}

	// Check for due date mismatch
	eventTime, err := time.Parse(time.RFC3339, event.Start.DateTime)
	if err != nil {
		return false, "", err
	}

	if !eventTime.Equal(task.Deadline) {
		log.Printf("task: %s, event: %s time needs update\n", task.Description, event.Summary)
		return true, NEEDS_UPDATE_DUE, nil
	}

	return false, "", nil
}

func ConvertTaskToCalendarEvent(task *model.Task) (*calendar.Event, error) {
	if task == nil {
		return nil, fmt.Errorf("could not convert nil Task")
	}

	if task.Deadline.IsZero() {
		return nil, fmt.Errorf("could not sync Task without due date: task id %s\n", task.ID)
	}

	var eventSummary string
	var eventStatus string

	switch task.Status {
	case "pending":
		eventSummary = task.Description
		eventStatus = "confirmed"
	case "completed":
		eventSummary = fmt.Sprintf("✅ %s", task.Description)
		eventStatus = "confirmed"
	case "deleted":
		eventSummary = fmt.Sprintf("❌ %s", task.Description)
		eventStatus = "cancelled"
	default:
		eventSummary = task.Description
		eventStatus = "confirmed"
	}

	defaultDuration := 30 * time.Minute
	eventDescription := fmt.Sprintf("Source: %s, ID: %s, Status: %s", task.Source, task.ID, task.Status)

	event := &calendar.Event{
		Summary: eventSummary,
		Status:  eventStatus,
		Start: &calendar.EventDateTime{
			DateTime: task.Deadline.UTC().Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: task.Deadline.UTC().Add(defaultDuration).Format(time.RFC3339),
		},
		Description: eventDescription,
	}

	return event, nil
}

// GetTaskIDFromEventDescription parses the task ID from the event description.
func GetTaskIDFromEventDescription(description string) (string, bool) {
	re := regexp.MustCompile(`ID: ([a-f0-9\-]+)`)
	matches := re.FindStringSubmatch(description)
	if len(matches) > 1 {
		return matches[1], true
	}
	return "", false
}