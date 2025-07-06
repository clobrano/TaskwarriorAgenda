package client

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/api/calendar/v3"
)

const (
	EVENT_COMPLETED_PREFIX = ""
	EVENT_DELETED_PREFIX   = "󰜺"
)

// Client provides methods to interact with the Google Calendar API.
type Client struct {
	srv *calendar.Service // The authenticated Google Calendar service
}

// NewClient creates and returns a new Google Calendar client.
// It takes the authenticated *calendar.Service as input.
func NewClient(srv *calendar.Service) *Client {
	return &Client{srv: srv}
}

// GetCalendarIDByName fetches the ID of a calendar given its summary (name).
// It performs a case-insensitive search. Returns an error if not found.
func (c *Client) GetCalendarIDByName(ctx context.Context, calendarName string) (string, error) {
	calendarList, err := c.srv.CalendarList.List().Do()
	if err != nil {
		return "", fmt.Errorf("unable to retrieve calendar list: %w", err)
	}

	for _, item := range calendarList.Items {
		// Use strings.EqualFold for case-insensitive comparison
		if strings.EqualFold(item.Summary, calendarName) {
			return item.Id, nil
		}
	}

	return "", fmt.Errorf("calendar with name '%s' not found", calendarName)
}

// ListEvents fetches events from a specified calendar within a given time range.
// It allows filtering by calendar ID and time.
func (c *Client) ListEvents(ctx context.Context, calendarID string, fromDate, toDate time.Time, maxResults int64) ([]*calendar.Event, error) {
	eventsCall := c.srv.Events.List(calendarID).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(fromDate.Format(time.RFC3339)).
		TimeMax(toDate.Format(time.RFC3339))

	if maxResults > 0 {
		eventsCall = eventsCall.MaxResults(maxResults)
	}

	events, err := eventsCall.Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve events from calendar '%s': %w", calendarID, err)
	}

	if len(events.Items) == 0 {
		log.Printf("No events found for calendar '%s'.", calendarID)
		return []*calendar.Event{}, nil
	}

	return events.Items, nil
}

// CreateEvent creates a new event in the specified calendar, linking it to a Taskwarrior UUID.
func (c *Client) CreateEvent(ctx context.Context, calendarID string, event *calendar.Event) error {
	createdEvent, err := c.srv.Events.Insert(calendarID, event).Do()
	if err != nil {
		return fmt.Errorf("unable to create event in calendar '%s': %w", event.Summary, err)
	}
	log.Printf("Google Calendar event created: '%s' %v", createdEvent.Summary, createdEvent.Start.DateTime)
	return nil
}

// UpdateEvent updates an existing event in the specified calendar.
func (c *Client) UpdateEvent(ctx context.Context, calendarID string, event *calendar.Event) error {
	if event.Id == "" {
		return fmt.Errorf("could not update event '%s': no event ID", event.Summary)
	}

	updatedEvent, err := c.srv.Events.Update(calendarID, event.Id, event).Do()
	if err != nil {
		return fmt.Errorf("unable to update event '%s' in calendar '%s': %w", event.Summary, calendarID, err)
	}
	log.Printf("Google Calendar event updated: '%s' %v", updatedEvent.Summary, updatedEvent.Start.DateTime)
	return nil
}
