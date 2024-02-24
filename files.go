package main

import (
	"io/ioutil"
	"strings"
)

func readLoggedEvents() []string {
	contents, err := ioutil.ReadFile("loggedEvents.csv")
	if err != nil {
		return []string{}
	}
	return strings.Split(string(contents), ",")
}

func saveLoggedEvents(events []string) {
	contents := strings.Join(events, ",")
	ioutil.WriteFile("loggedEvents.csv", []byte(contents), 0644)
}
