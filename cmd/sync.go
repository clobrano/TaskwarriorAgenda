/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"log"
	"strings"

	"github.com/clobrano/TaskwarriorAgenda/pkg/auth"
	client "github.com/clobrano/TaskwarriorAgenda/pkg/google"
	"github.com/clobrano/TaskwarriorAgenda/pkg/taskwarrior"
	"github.com/clobrano/TaskwarriorAgenda/pkg/util"
	"github.com/spf13/cobra"
)

var (
	calendarName      string
	taskWarriorFilter []string
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize tasks between Taskwarrior and Google calendar",
	Long:  `TODO`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		log.Println("Initializing Google Calendar service...")
		calendarService, err := auth.GetCalendarService(ctx)
		if err != nil {
			log.Fatalf("Unable to get Google Calendar service: %v", err)
		}

		t := taskwarrior.NewClient()
		//filter := []string{"+PENDING", "+rem"}
		tasks, err := t.GetTasks(taskWarriorFilter)
		if err != nil {
			log.Println(err)
			return
		}

		if len(tasks) == 0 {
			log.Printf("Filter %v returned 0 taskwarrior tasks", taskWarriorFilter)
			return
		} else {
			log.Printf("Filter %v returned %d taskwarrior tasks", taskWarriorFilter, len(tasks))
		}

		c := client.NewClient(calendarService)
		calendarID, err := c.GetCalendarIDByName(ctx, calendarName)
		if err != nil {
			log.Printf("could not get Calendar ID from calendar name '%s': %v\n", "To-do", err)
			return
		}

		fromDate, toDate := util.GetDateRange(tasks)
		log.Printf("Getting data from %v (year %d) to %v\n", fromDate, fromDate.Year(), toDate)
		events, err := c.ListEvents(ctx, calendarID, fromDate, toDate, 0)

		needUpdate := []taskwarrior.Task{}
		needCreation := []taskwarrior.Task{}
		if err != nil {
			log.Printf("could not get events: error %v\n", err)
		} else {
			log.Printf("Application finished successfully. Found %d events\n", len(events))
			for _, t := range tasks {
				foundMatchingEvent := false
				for _, e := range events {
					if !strings.Contains(e.Description, t.UUID) {
						continue
					}

					foundMatchingEvent = true
					got, err := util.EventNeedsUpdate(&t, e)
					if err != nil {
						log.Printf("could not compare task uuid %s with its calendar event: error %v", t.UUID, err)
					} else if got {
						needUpdate = append(needUpdate, t)
					}
				}
				if !foundMatchingEvent {
					needCreation = append(needCreation, t)
				}
			}
		}

		for _, t := range needUpdate {
			if newEvent, err := util.ConvertTaskwarriorTaskToCalendarEvent(&t); err != nil {
				log.Printf("could not convert Task '%s' into event: %v\n", t.Description, err)
			} else {
				err = c.UpdateEvent(ctx, calendarID, newEvent)
				if err != nil {
					log.Printf("could not update event for task '%s': %v\n", t.Description, err)
				}
			}
		}

		for _, t := range needCreation {
			if newEvent, err := util.ConvertTaskwarriorTaskToCalendarEvent(&t); err != nil {
				log.Printf("could not convert Task '%s' into event: %v\n", t.Description, err)
			} else {
				err = c.CreateEvent(ctx, calendarID, newEvent)
				if err != nil {
					log.Printf("could not create Event '%s': %v\n", t.Description, err)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// syncCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// syncCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	syncCmd.Flags().StringVarP(&calendarName, "calendar", "c", "", "Name of the calendar to operate on")
	syncCmd.MarkFlagRequired("calendar")
	syncCmd.Flags().StringArrayVarP(&taskWarriorFilter, "filter", "f", []string{"+PENDING"}, "filter to select taskwarrior tasks to syncronize")
}
