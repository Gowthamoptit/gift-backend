package notifyslack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type SlackMessage struct {
	Text string `json:"text"`
}

// Function to send message to Slack
func sendToSlack(message string) error {
	webhookURL := "https://hooks.slack.com/services/xxxxxx/xxxxxx/xxxxxxxx" // Replace with your Slack Webhook URL

	slackMessage := SlackMessage{
		Text: message,
	}

	// Convert message to JSON
	jsonData, err := json.Marshal(slackMessage)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	// Send HTTP POST request to Slack Webhook URL
	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Log the status code and response body for debugging
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("error sending message to Slack: %s, body: %s", resp.Status, string(body))
	}

	// Optionally log the success status (optional, just for debugging)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("Slack response: %s\n", string(body))

	return nil
}
