# Taskwarrior Agenda

This tool synchronizes tasks from Taskwarrior to a Google Calendar. It helps you visualize your tasks in a calendar view, making it easier to plan your day and stay organized.

## Features

*   **Task Synchronization:** Automatically syncs tasks from Taskwarrior to a specified Google Calendar.
*   **Date Range Handling:**  Correctly handles tasks with due dates and scheduled dates.
*   **Status Updates:** Updates calendar events based on Taskwarrior task status (e.g., completed, deleted).
*   **Customizable:**  Allows you to configure the Google Calendar to use and the filters for selecting tasks.

## Prerequisites

*   [Taskwarrior](https://taskwarrior.org/) installed and configured.
*   Go installed (version 1.16 or later).
*   A Google Cloud project with the Google Calendar API enabled.
*   Credentials file (`credentials.json`) for your Google Cloud project.

## Installation

```bash
go install github.com/clobrano/TaskwarriorAgenda
```

## Configuration

1.  **Google Cloud Credentials:**
    * Download the `credentials.json` file from your Google Cloud project.
    * Place the `credentials.json` file in `~/.config/taskwarrior-agenda` directory
    * Run `TaskwarriorAgenda auth` and follow the instructions. The authentication file `token.json` will be stored in the same `~/.config/taskwarrior-agenda` directory.
      * To refresh the authentication token, delete the old `token.json` before running `auth` again.

2.  **Google Calendar API:**
    * Ensure the Google Calendar API is enabled for your project.

3.  **Configuration Files:**
    * The application uses the `credentials.json` file to get the `token.json`, both files are stored by default in `~/.config/taskwarrior-agenda` directory


## Usage

```bash
Usage:
  TaskwarriorAgenda [command]

Available Commands:
  auth        Authenticate with Google Calendar API
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  sync        Synchronize tasks between Taskwarrior and Google calendar

Flags:
  -h, --help     help for TaskwarriorAgenda
  -t, --toggle   Help message for toggle

Use "TaskwarriorAgenda [command] --help" for more information about a command.
```

### Example

```bash
# Sync only tasks with tag "reminder", not deleted and modified in the last 7 days
./TaskwarriorAgenda sync --calendar "To-do" --filter "+reminder -DELETED modified.after=-7d"
```

## Contributing

Contributions are welcome! Please submit a pull request with your changes.
