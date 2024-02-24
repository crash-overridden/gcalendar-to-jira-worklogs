package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"github.com/pkg/browser"
)

func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}

	return config.Client(context.Background(), tok)
}

func getCodeUsingHttpServer(authCode chan string) {
	srv := &http.Server{Addr: ":8000"}
	success := false
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "<html> <head> </head> <body> <h3>Thats all, you can close that page</h3> </body> </html>")
		authCode <- req.URL.Query().Get("code")
		success = true
		go func() {
			time.Sleep(5 * time.Second)
			srv.Shutdown(context.TODO())
		}()
	})
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			if !success {
				fmt.Println("Smth went wrong with http server")
			}
		}
	}()
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	browser.OpenURL(authURL)

	// fmt.Printf("Go to the following link in your browser then type the "+
	// 	"authorization code: \n%v\n", authURL)

	authCodeChannel := make(chan string)
	getCodeUsingHttpServer(authCodeChannel)

	authCode := <-authCodeChannel

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

type UserInfo struct {
	Sub           string `json:"sub"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Hd            string `json:"hd"`
}

func getUserEmail(client *http.Client) string {
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")

	if err != nil {
		log.Fatalf("Request for user email error: %v \n", err)
	}

	defer resp.Body.Close()

	// logResponse(resp)
	body, err := ioutil.ReadAll(resp.Body)

	var result UserInfo
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalf("Can not unmarshal JSON: %v \n", err)
	}

	return result.Email
}

func authenticateGoogle() (*calendar.Service, *http.Client) {
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope, "https://www.googleapis.com/auth/userinfo.email")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	return srv, client
}

func getCalendarEvents(srv *calendar.Service, tMin time.Time, tMax time.Time) (*calendar.Events, error) {
	tmax := tMax.Format(time.RFC3339)
	tmin := tMin.Format(time.RFC3339)

	events, err := srv.Events.List("primary").ShowDeleted(false).
		SingleEvents(true).TimeMin(tmin).TimeMax(tmax).OrderBy("startTime").Do()

	return events, err
}

func findJiraTicket(event *calendar.Event) *string {
	r := regexp.MustCompile(`[A-Z0-9]{2,}-\d+`)
	matches := r.FindAllString(event.Description, -1)
	matches = append(matches, r.FindAllString(event.Summary, -1)...)

	if len(matches) > 0 {
		return &matches[0]
	} else {
		return nil
	}
}

func eventDurationInSeconds(startDate string, endDate string) int {
	startTime, _ := time.Parse(time.RFC3339, startDate)
	endTime, _ := time.Parse(time.RFC3339, endDate)

	return int(endTime.Sub(startTime).Seconds())
}
