package main

import (
	. "auth0-logs/shared"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	from, err := parseFromArgument(os.Args)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}

	db, err := connectToPostgres()
	if err != nil {
		fmt.Println("Error connecting to PostgreSQL:", err)
		os.Exit(1)
	}
	defer db.Close()

	token, err := GetAuth0Token()
	if err != nil {
		fmt.Println("Error obtaining Auth0 token:", err)
		os.Exit(1)
	}

	fmt.Println("Fetching logs...")
	var url string

	for {
		logs, nextURL, err := fetchLogs(token, url, from, 100)
		if err != nil {
			fmt.Printf("Error fetching logs [from=%s, url=%s]: %s\n", from, url, err)
			os.Exit(1)
		}

		if len(logs) == 0 {
			break
		}

		for _, log := range logs {
			err = insertLog(db, log)
			if err != nil {
				fmt.Printf("Error inserting log %s: %v\n", log.LogID, err)
				os.Exit(1)
			}
		}

		url = nextURL
		from = ""
	}

	fmt.Println("Completed")
}

func parseFromArgument(args []string) (string, error) {
	if len(args) < 2 {
		return "", errors.New("log ID not provided. Usage: go run main.go <log-id>")
	}

	logID := args[1]
	if !regexp.MustCompile("^[0-9]+$").MatchString(logID) {
		return "", errors.New("invalid log ID. It should only contain numbers")
	}

	return logID, nil
}

func connectToPostgres() (*sql.DB, error) {
	fmt.Print("\nConnecting to Postgres... ")
	connStr := "dbname=public host=localhost port=5432 sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("Done!")
	return db, nil
}

func insertLog(db *sql.DB, log LogEntry) error {
	query := `
		INSERT INTO public.logs (log_id, date, type, size)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (log_id) DO NOTHING
	`

	_, err := db.Exec(query, log.LogID, log.Date, log.Type, log.Size)
	if err != nil {
		return fmt.Errorf("error inserting log %s: %v", log.LogID, err)
	}

	return nil
}

func fetchLogs(token, url, from string, take int) ([]LogEntry, string, error) {
	var (
		req *http.Request
		err error
	)

	if len(url) > 0 {
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, "", err
		}

	} else if len(from) > 0 {
		req, err = http.NewRequest("GET", LogsEndpoint, nil)
		if err != nil {
			return nil, "", err
		}

		q := req.URL.Query()
		q.Add("from", from)
		q.Add("take", strconv.Itoa(take))
		req.URL.RawQuery = q.Encode()

	} else {
		return nil, "", errors.New("invalid parameters")
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	//fmt.Printf("Fetching logs with %s", req.URL.String())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	nextURL, err := parseNextURL(resp.Header)
	if err != nil {
		return nil, "", err
	}

	body, _ := io.ReadAll(resp.Body)
	logs, err := BytesToLogEntries(body)
	if err != nil {
		fmt.Printf("Error calling bytesToLogEntries: %s, response body: %s\n", err.Error(), string(body))
	}

	//fmt.Printf(" > %d logs\n", len(logs))
	fmt.Printf("%d, ", len(logs))

	return logs, nextURL, nil
}

func parseNextURL(headers http.Header) (string, error) {
	linkHeader := headers.Get("Link")
	if linkHeader == "" {
		return "", errors.New("header 'Link' not found")
	}

	links := strings.Split(linkHeader, ",")
	for _, link := range links {
		if strings.Contains(link, "rel=\"next\"") {
			parts := strings.Split(link, ";")
			url := strings.Trim(parts[0], " <>")
			return url, nil
		}
	}
	return "", errors.New("'next' link not found")
}
