package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const LogsEndpoint = "https://sothebys.auth0.com/api/v2/logs"

type LogEntry struct {
	Date  time.Time `json:"date"`
	LogID string    `json:"log_id"`
	Type  string    `json:"type"`
	Size  int       `json:"-"`
}

func GetAuth0Token() (string, error) {
	data := map[string]string{
		"client_id":     "RSU2FFQI56uAljNGo5SSnc6uMr7DojYl",
		"client_secret": "3VG0qgsV3dBY5EaabeB0YrqCZrJqJE9YnFl6u6xM-cz3WZukyF2uY80DheK5CRjB",
		"audience":      "https://sothebys.auth0.com/api/v2/",
		"grant_type":    "client_credentials",
		"scope":         "read:stats",
	}
	jsonData, _ := json.Marshal(data)

	resp, err := http.Post("https://sothebys.auth0.com/oauth/token", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to obtain token: %s, %s", resp.Status, string(body))
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", err
	}

	return tokenResponse.AccessToken, nil
}

func BytesToLogEntries(b []byte) ([]LogEntry, error) {
	var logs []LogEntry
	if err := json.Unmarshal(b, &logs); err != nil {
		return nil, err
	}

	for i, log := range logs {
		logData, err := json.Marshal(log)
		if err != nil {
			return nil, err
		}
		logs[i].Size = len(logData)
	}

	return logs, nil
}
