package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// Exit codes used by NZBGet
const POSTPROCESS_SUCCESS = 93
const POSTPROCESS_ERROR = 94
const POSTPROCESS_NONE = 95

func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func sendSlackMessage(msg string, token string, channel string) bool {
	url := "https://slack.com/api/chat.postMessage"

	msqJsonMap := map[string]string{
		"username":     "NZBGet",
		"icon_emoji":   ":seal:",
		"unfurl_links": "false",
		"channel":      channel,
		"text":         msg,
	}
	msgJson, _ := json.Marshal(msqJsonMap)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(msgJson))
	if err != nil {
		fmt.Printf("[ERROR] error creating http request: %s", err.Error())
		return false
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("[ERROR] error sending Slack message: %s", err.Error())
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("[ERROR] error sending Slack message: got status code: %s", resp.Status)
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[ERROR] error reading Slack responce body: %s", err.Error())
		return false
	}

	var respJsonMap map[string]interface{}
	if err := json.Unmarshal(body, &respJsonMap); err != nil {
		fmt.Printf("[ERROR] error parsing Slack responce body JSON: %s", err.Error())
		return false
	}

	if status, ok := respJsonMap["ok"].(bool); ok {
		if !status {
			log.Printf("[ERROR] error sending Slack message: %v", respJsonMap["error"])
			return false
		}
	}

	return true
}

func main() {
	// get settings from environment
	token := getEnv("NZBPO_SLACKTOKEN", "")
	channel := getEnv("NZBPO_SLACKCHANNEL", "")
	sendNotification := getEnv("NZBPO_SENDNOTIFICATION", "")

	// check settings
	if token == "" || channel == "" {
		fmt.Println("[ERROR] Please set up token and channel")
		os.Exit(POSTPROCESS_ERROR)
	}

	// list all environment variables and their values if debug enabled
	debug := getEnv("NZBOP_SLACKNOTIFY_DEBUG", "")
	if debug == "Yes" {
		fmt.Println("[DETAIL] environment:")
		for _, env := range os.Environ() {
			fmt.Printf("[DETAIL] %s\n", env)
		}
	}

	// process commands
	command := getEnv("NZBCP_COMMAND", "")
	if command == "ConnectionTest" {
		if sendSlackMessage("Connection test", token, channel) {
			os.Exit(POSTPROCESS_SUCCESS)
		} else {
			os.Exit(POSTPROCESS_ERROR)
		}
	}

	// process script
	status := getEnv("NZBPP_STATUS", "")
	send := true
	if status == "SUCCESS/ALL" && sendNotification == "OnFailure" {
		send = false
	}

	if send {
		nzbname := getEnv("NZBPP_NZBNAME", "")
		msg := fmt.Sprintf("Download of \"%s\" completed.\nStatus: %s", nzbname, status)
		if sendSlackMessage(msg, token, channel) {
			os.Exit(POSTPROCESS_SUCCESS)
		} else {
			os.Exit(POSTPROCESS_ERROR)
		}
	}

	os.Exit(POSTPROCESS_SUCCESS)
}
