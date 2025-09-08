package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/clobrano/TaskwarriorAgenda/pkg/google"
	"github.com/clobrano/TaskwarriorAgenda/pkg/model"
	"github.com/clobrano/TaskwarriorAgenda/pkg/orgmode"
	"github.com/clobrano/TaskwarriorAgenda/pkg/taskwarrior"
	"github.com/clobrano/TaskwarriorAgenda/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize tasks to Google calendar",
	Long:  `Synchronize tasks from Taskwarrior or an Org-mode file to Google calendar.`,
	Run: func(cmd *cobra.Command, args []string) {
		calendar, _ := cmd.Flags().GetString("calendar")
		source, _ := cmd.Flags().GetString("source")
		filter, _ := cmd.Flags().GetString("filter")

		var tasks []model.Task
		var err error

		switch source {
		case "orgmode":
			// Get the list of Org-mode files from the configuration
			files := viper.GetStringSlice("orgmode_files")
			if len(files) == 0 {
				log.Fatal("Error: no Org-mode files specified in the configuration file")
			}
			tasks, err = orgmode.ParseFiles(files)
			if err != nil {
				log.Fatalf("Error parsing Org-mode files: %v", err)
			}
			if filter != "" {
				tasks = orgmode.FilterTasks(tasks, filter)
			}
		case "taskwarrior":
			client := taskwarrior.NewClient()
			twTasks, err := client.GetTasks(strings.Split(filter, " "))
			if err != nil {
				log.Fatalf("Error getting tasks from Taskwarrior: %v", err)
			}
			// Convert taskwarrior.Task to model.Task
			for _, t := range twTasks {
				var deadline time.Time
				if t.Due != nil {
					deadline = t.Due.Time
				}
				tasks = append(tasks, model.Task{
					ID:          t.UUID,
					Description: t.Description,
					Deadline:    deadline,
					Status:      t.Status,
					Source:      "taskwarrior",
				})
			}
		default:
			log.Fatalf("Error: invalid source '%s'. Please use 'taskwarrior' or 'orgmode'", source)
		}

		sync(calendar, tasks)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().String("calendar", "Tasks", "Google Calendar name to sync with")
	syncCmd.Flags().String("source", "", "Source of tasks (taskwarrior or orgmode)")
	syncCmd.MarkFlagRequired("source")
	syncCmd.Flags().String("filter", "", "Filter to apply to the tasks")
}

func sync(calendarName string, tasks []model.Task) {
	client, err := google.NewClient(calendarName)
	if err != nil {
		log.Fatalf("Error creating Google Calendar client: %v", err)
	}

	// Create a map of task IDs for efficient lookup
	taskMap := make(map[string]bool)
	for _, task := range tasks {
		taskMap[task.ID] = true
	}

	// TODO: improve orphaned events management. For now it is disabled
	// the problem is that looking for all the events will take more and more
	// time in the future, since we'll have more and more events in history.
	// The solution is to look only in Orgmode.md not in Orgmode_archive.md. However,
	// as of today implementation, removing the latter from the watch list, will make a
	// bunch of events form the Calendar "orphaned" and so deleted.

	if false {
		// Fetch recent events to check for orphans
		events, err := client.ListEvents(time.Now().Add(-30 * 24 * time.Hour))
		if err != nil {
			log.Fatalf("Error fetching calendar events: %v", err)
		}

		// Delete orphaned events
		for _, event := range events {
			taskID, found := util.GetTaskIDFromEventDescription(event.Description)
			if found && !taskMap[taskID] {
				log.Printf("Deleting orphaned event for task '%s' ID %s", event.Description, taskID)
				err := client.DeleteEvent(event.Id)
				if err != nil {
					log.Printf("Error deleting event: %v", err)
				}
			}
		}
	}

	// Sync current tasks
	for _, task := range tasks {
		fmt.Printf("Syncing task: '%s', uuid: %s, due: %s\n", task.Description, task.ID, task.Deadline)
		_, err := client.SyncEvent(task)
		if err != nil {
			fmt.Printf("Error syncing event for task %s: %v\n", task.Description, err)
		}
	}
}
