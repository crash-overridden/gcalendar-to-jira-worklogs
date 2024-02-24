package main

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

func clearOldLoggedEvents(loggedEvents []string) []string {
	cleanLoggedEvents := []string{}

	for _, loggedEvent := range loggedEvents {
		loggedEventTime, err := time.Parse(time.RFC3339, strings.Split(loggedEvent, "_")[0])
		if err == nil && loggedEventTime.After(time.Now().Add(-48*time.Hour)) {
			cleanLoggedEvents = append(cleanLoggedEvents, loggedEvent)
		}
	}

	return cleanLoggedEvents
}

func main() {
	calendarService, httpClient := authenticateGoogle()
	loggedEvents := clearOldLoggedEvents(readLoggedEvents())
	email := getUserEmail(httpClient)

	for true {
		events, err := getCalendarEvents(calendarService, time.Now().Add(-28*time.Hour), time.Now())
		if err != nil {
			fmt.Printf("Unable to retrieve events: %v", err)
		} else {
			if len(events.Items) == 0 {
				fmt.Println("No events found.")
			} else {
				for _, event := range events.Items {
					// fmt.Printf("%v \n", event.Summary)
					ticket := findJiraTicket(event)

					if ticket == nil {
						continue
					}

					for _, attendee := range event.Attendees {
						if attendee.Email == email && attendee.ResponseStatus == "accepted" && !slices.Contains(loggedEvents, event.Start.DateTime+"_"+*ticket) {
							loggedEvents = append(loggedEvents, (event.Start.DateTime + "_" + *ticket))
							logWorkOnJira(*ticket, eventDurationInSeconds(event.Start.DateTime, event.End.DateTime), event.Start.DateTime, event.Summary)
						}
					}

					if len(event.Attendees) == 0 && !slices.Contains(loggedEvents, event.Start.DateTime+"_"+*ticket) {
						loggedEvents = append(loggedEvents, (event.Start.DateTime + "_" + *ticket))
						logWorkOnJira(*ticket, eventDurationInSeconds(event.Start.DateTime, event.End.DateTime), event.Start.DateTime, event.Summary)
					}
				}
			}
		}

		saveLoggedEvents(loggedEvents)

		time.Sleep(15 * time.Minute)
	}
}
