package main

import (
	. "auth0-logs/shared"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const validDateFormats = `Please provide a date-time argument in one of the following formats:

 - yyyy-MM-dd
 - yyyy-MM-dd HH:mm
 - yyyy-MM-dd HH:mm:ss
 `

func main() {
	if len(os.Args) < 2 {
		fmt.Println(validDateFormats)
		os.Exit(1)
	}

	dateTimeStr := os.Args[1]
	dateTime, err := parseDateArgument(dateTimeStr)
	if err != nil {
		fmt.Println("Error parsing date-time:", err)
		os.Exit((1))
	}

	token, err := GetAuth0Token()
	if err != nil {
		fmt.Println("Error obtaining Auth0 token:", err)
		os.Exit((1))
	}

	logID, err := fetchFirstLogID(token, dateTime)
	if err != nil {
		fmt.Println("Error fetching first log ID:", err)
		os.Exit((1))
	}

	fmt.Println("LOG ID: ", logID)
}

func fetchFirstLogID(token string, dateTime time.Time) (string, error) {

	endTime := dateTime.Add(12 * time.Hour)

	startStr := dateTime.Format("2006-01-02T15:04:05")
	endStr := endTime.Format("2006-01-02T15:04:05")

	fmt.Printf("Fetching first log ID from %s\n", startStr)
	req, err := http.NewRequest("GET", LogsEndpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	q := req.URL.Query()
	q.Add("q", fmt.Sprintf("date:[%s TO %s]", startStr, endStr))
	q.Add("per_page", "1")

	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	logs, err := BytesToLogEntries(body)
	if err != nil {
		fmt.Printf("Error calling BytesToLogEntries: %s, response body: %s", err.Error(), string(body))
		os.Exit((1))
	}

	if len(logs) == 0 {
		return "", errors.New("first log not found")
	}

	return logs[0].LogID, nil
}

// parseDateArgument parses a date-time string passed as an argument
// and returns a time.Time value. It supports three formats:
// 1. "2024-11-14" (date only)
// 2. "2024-11-14 12:12" (date and hour)
// 3. "2024-11-14 12:12:00" (full date and time)
func parseDateArgument(dateTimeStr string) (time.Time, error) {
	if t, err := time.Parse("2006-01-02", dateTimeStr); err == nil {
		return t, nil
	}

	if t, err := time.Parse("2006-01-02 15:04", dateTimeStr); err == nil {
		return t, nil
	}

	if t, err := time.Parse("2006-01-02 15:04:05", dateTimeStr); err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("invalid date-time format. %s", validDateFormats)
}
