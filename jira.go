package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
)

type Config struct {
	Jira struct {
		Email    string `json:"email"`
		Token string `json:"token"`
	} `json:"jira"`
}
type Payload struct {
	TimeSpentSeconds int     `json:"timeSpentSeconds"`
	Comment          Comment `json:"comment"`
	Started          string  `json:"started"`
}
type Content struct {
	Text string `json:"text"`
	Type string `json:"type"`
}
type Contents struct {
	Type    string    `json:"type"`
	Content []Content `json:"content"`
}
type Comment struct {
	Type     string     `json:"type"`
	Version  int        `json:"version"`
	Contents []Contents `json:"content"`
}

func loadConfig(path string) (*Config, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	var config Config
	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func prepareWorklogJson(ticket string, durationInSeconds int, startDate string, comment string) *bytes.Reader {
	data := Payload{
		TimeSpentSeconds: durationInSeconds,
		Comment: Comment{
			Type:    "doc",
			Version: 1,
			Contents: []Contents{
				{
					Type: "paragraph",
					Content: []Content{
						{
							Text: comment,
							Type: "text",
						},
					},
				},
			},
		},
		Started: toJiraDate(startDate),
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Parsing error: %v \n", err)
	}
	body := bytes.NewReader(payloadBytes)

	return body
}

func toJiraDate(startDate string) string {
	startTime, _ := time.Parse(time.RFC3339, startDate)

	return startTime.Format("2006-01-02T15:04:05.000+0000")
}

func logRequest(req *http.Request) {
	reqDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Fatalf("Dump request error: %v\n", err)
	}

	fmt.Printf("\n\n----------HTTP REQUEST----------\n%s", string(reqDump))
}

func logResponse(resp *http.Response) {
	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n\n----------HTTP RESPONSE----------\n%s", string(respDump))
}

func logWorkOnJira(ticket string, durationInSeconds int, startDate string, comment string) {
	body := prepareWorklogJson(ticket, durationInSeconds, startDate, comment)

	req, err := http.NewRequest("POST", "https://thrnd.atlassian.net/rest/api/3/issue/"+ticket+"/worklog", body)
	if err != nil {
		fmt.Printf("Jira request error: %v \n", err)
		return
	}

	config, err := loadConfig("config.json")
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	req.SetBasicAuth(config.Jira.Email, config.Jira.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	// logRequest(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Request to jira error: %v \n", err)
		return
	}

	defer resp.Body.Close()

	// logResponse(resp)

	if resp.StatusCode == 201 {
		fmt.Printf("Logged work for [%v], %v, %v minutes, %v \n", ticket, startDate, (durationInSeconds / 60), comment)
	} else {
		fmt.Printf("Jira error: %v \n", resp.StatusCode)
	}
}
