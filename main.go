package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	// Ask BotFather
	TelegramBotToken = ""
	// Invite this bot to get your group ChatID, Chat ID Bot (@chat_id_echo_bot) / https://web.telegram.org/a/#1513323938
	ChatID = ""
)

type GitLabEvent struct {
	ObjectKind   string `json:"object_kind"`
	UserUsername string `json:"user_username"`
	Ref          string `json:"ref"`
	Project      struct {
		WebURL        string `json:"web_url"`
		Name          string `json:"name"`
		DefaultBranch string `json:"default_branch"`
	} `json:"project"`
	ObjectAttributes struct {
		URL    string `json:"url"`
		Title  string `json:"title"`
		Action string `json:"action"`
	} `json:"object_attributes"`
	User struct {
		Username string `json:"username"`
	} `json:"user"`
	TotalCommitCount int `json:"total_commits_count"`
}

func main() {
	http.HandleFunc("/", handleWebhook)
	log.Println("Listening at 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Error parsing response body: %s", err)
		log.Println(err)
		return
	}
	fmt.Println(string(b))

	var event GitLabEvent
	err = json.Unmarshal(b, &event)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error decoding JSON: %s", err)
		return
	}

	message := generateMessage(event)
	if message != "" {
		sendTelegramMessage(message)
	}

	w.WriteHeader(http.StatusOK)
}

func generateMessage(event GitLabEvent) string {
	switch event.ObjectKind {
	case "push":
		if event.TotalCommitCount > 0 {
			branchName := strings.TrimPrefix(event.Ref, "refs/heads/")
			branchLink := fmt.Sprintf("%s/-/tree/%s", event.Project.WebURL, url.QueryEscape(branchName))
			return fmt.Sprintf("🔨 New push by %s to %s:\n\nTarget branch: %s\nLink: %s",
				event.UserUsername, event.Project.Name, event.Ref, branchLink)
		}
	case "merge_request":
		if event.ObjectAttributes.Action == "approved" {
			return fmt.Sprintf("👍 This merge request get APPROVED by %s:\n\nTitle: %s\nLink: %s",
				event.User.Username, event.ObjectAttributes.Title, event.ObjectAttributes.URL)
		} else if event.ObjectAttributes.Action == "unapproved" {
			return fmt.Sprintf("👎 This merge request get UNAPPROVED by %s:\n\nTitle: %s\nLink: %s",
				event.User.Username, event.ObjectAttributes.Title, event.ObjectAttributes.URL)
		} else if event.ObjectAttributes.Action == "open" || event.ObjectAttributes.Action == "reopen" {
			return fmt.Sprintf("🔥 New merge request opened by %s:\n\nTitle: %s\nLink: %s",
				event.User.Username, event.ObjectAttributes.Title, event.ObjectAttributes.URL)
		} else if event.ObjectAttributes.Action == "close" {
			return fmt.Sprintf("❌ Merge request get closed by %s:\n\nTitle: %s\nLink: %s",
				event.User.Username, event.ObjectAttributes.Title, event.ObjectAttributes.URL)
		} else if event.ObjectAttributes.Action == "merge" {
			return fmt.Sprintf("🎉 Merge request get MERGED by %s:\n\nTitle: %s\nLink: %s",
				event.User.Username, event.ObjectAttributes.Title, event.ObjectAttributes.URL)
		}
	case "note":
		return fmt.Sprintf("💬 New comment by %s:\n\nLink: %s",
			event.User.Username, event.ObjectAttributes.URL)
	default:
		// Unknown type, do not send
		return ""
	}
	return ""
}

func sendTelegramMessage(message string) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", TelegramBotToken)
	data := url.Values{
		"chat_id": {ChatID},
		"text":    {message},
	}

	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		log.Printf("Error sending message to Telegram: %s", err)
		return
	}
	defer resp.Body.Close()
}
