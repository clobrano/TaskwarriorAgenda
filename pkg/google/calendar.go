package google

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/clobrano/TaskwarriorAgenda/pkg/model"
	"github.com/clobrano/TaskwarriorAgenda/pkg/util"
	"google.golang.org/api/calendar/v3"
)

// CalendarClient is a Google Calendar API client.
type CalendarClient struct {
	srv        *calendar.Service
	calendarID string
}

// NewCalendarClient creates a new Google Calendar client.
func NewCalendarClient(srv *calendar.Service, calendarID string) *CalendarClient {
	return &CalendarClient{srv: srv, calendarID: calendarID}
}

// SyncEvent creates a new event or updates an existing one.
func (c *CalendarClient) SyncEvent(task model.Task) (*calendar.Event, error) {
	event, err := util.ConvertTaskToCalendarEvent(&task)
	if err != nil {
		return nil, err
	}

	// Search for existing event
	events, err := c.ListEvents(time.Now().Add(-30 * 24 * time.Hour))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve events from calendar: %w", err)
	}

	for _, existingEvent := range events {
		if strings.Contains(existingEvent.Description, fmt.Sprintf("ID: %s", task.ID)) {
			needsUpdate, _, err := util.EventNeedsUpdate(&task, existingEvent)
			if err != nil {
				log.Printf("could not compare task with its calendar event: %v", err)
				continue
			}
			if needsUpdate {
				log.Printf("Updating event for task: %s", task.Description)
				return c.srv.Events.Update(c.calendarID, existingEvent.Id, event).Do()
			}
			log.Printf("Event for task %s is already up to date", task.Description)
			return existingEvent, nil
		}
	}

	log.Printf("Creating new event for task: %s", task.Description)
	return c.srv.Events.Insert(c.calendarID, event).Do()
}

// DeleteEvent deletes an event from the calendar.
func (c *CalendarClient) DeleteEvent(eventID string) error {
	return c.srv.Events.Delete(c.calendarID, eventID).Do()
}

// ListEvents fetches events from the calendar within a given time range.
func (c *CalendarClient) ListEvents(timeMin time.Time) ([]*calendar.Event, error) {
	events, err := c.srv.Events.List(c.calendarID).TimeMin(timeMin.Format(time.RFC3339)).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve events from calendar: %w", err)
	}
	return events.Items, nil
}
