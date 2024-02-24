# Google Calendar to JIRA Worklogs

This Go project synchronizes events from Google Calendar with worklogs in JIRA, facilitating an automated process for logging work hours. It is particularly beneficial for teams leveraging JIRA for project management alongside Google Calendar for event scheduling.

## Features

- **Automatic Synchronization**: Seamlessly syncs events from Google Calendar directly into JIRA as worklogs.
- **Configuration File**: Utilizes a `config.json` file for easy management of JIRA credentials and other settings.
- **OAuth2 Authentication**: Employs OAuth2 for secure access to Google Calendar and JIRA APIs, ensuring data safety.

## Prerequisites

- Go version 1.15 or later.
- Google Cloud Platform project with Calendar API enabled.
- JIRA account with permissions to create worklogs.

## Installation

1. Clone the repository to your local machine:
   ```
   git clone https://github.com/crash-overridden/gcalendar-to-jira-worklogs.git
   ```
2. Navigate into the project directory:
   ```
   cd gcalendar-to-jira-worklogs
   ```
3. Install the necessary dependencies:
   ```
   go mod tidy
   ```

## Configuration

Before running the application, you need to configure the `config.json` file with your JIRA credentials and Google Calendar settings. This includes setting up the `credentials.json` for Google Calendar API and specifying your JIRA email and password in the `config.json`.

## Usage

To run the application and start syncing your Google Calendar events with JIRA worklogs, execute:
```
go run main.go
```

## Contributing

Contributions are welcome! If you have any suggestions for improving this project, please feel free to make a pull request or open an issue.

## License

This project is licensed under the MIT License.