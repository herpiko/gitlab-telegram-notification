package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
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
	TotalCommitCount   int    `json:"total_commits_count"`
	BuildName          string `json:"build_name"`
	BuildStatus        string `json:"build_status"`
	BuildFailureReason string `json:"build_failure_reason"`
}

type GitlabTelegram struct {
	BotToken string
	ChatID   string
}

func (g *GitlabTelegram) HandleWebhook(w http.ResponseWriter, r *http.Request) {
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

	message := g.generateMessage(event)
	if message != "" {
		g.sendTelegramMessage(message)
	}

	w.WriteHeader(http.StatusOK)
}

func (g *GitlabTelegram) generateMessage(event GitLabEvent) string {
	switch event.ObjectKind {
	case "push":
		if event.TotalCommitCount > 0 {
			branchName := strings.TrimPrefix(event.Ref, "refs/heads/")
			branchLink := fmt.Sprintf("%s/-/tree/%s", event.Project.WebURL, url.QueryEscape(branchName))
			return fmt.Sprintf("🔨 New push by %s to %s.\nTarget branch: %s\n%s",
				event.UserUsername, event.Project.Name, event.Ref, branchLink)
		}
	case "merge_request":
		if event.ObjectAttributes.Action == "approved" {
			return fmt.Sprintf("👍 This merge request get APPROVED by %s: %s\n%s",
				event.User.Username, event.ObjectAttributes.Title, event.ObjectAttributes.URL)
		} else if event.ObjectAttributes.Action == "unapproved" {
			return fmt.Sprintf("👎 This merge request get UNAPPROVED by %s: %s\n%s",
				event.User.Username, event.ObjectAttributes.Title, event.ObjectAttributes.URL)
		} else if event.ObjectAttributes.Action == "open" || event.ObjectAttributes.Action == "reopen" {
			return fmt.Sprintf("🔥 New merge request opened by %s: %s\n%s",
				event.User.Username, event.ObjectAttributes.Title, event.ObjectAttributes.URL)
		} else if event.ObjectAttributes.Action == "close" {
			return fmt.Sprintf("❌ Merge request get closed by %s: %s\n%s",
				event.User.Username, event.ObjectAttributes.Title, event.ObjectAttributes.URL)
		} else if event.ObjectAttributes.Action == "merge" {
			return fmt.Sprintf("🎉 Merge request get MERGED by %s: %s\n%s",
				event.User.Username, event.ObjectAttributes.Title, event.ObjectAttributes.URL)
		}
	case "note":
		return fmt.Sprintf("💬 New comment by %s.\n%s",
			event.User.Username, event.ObjectAttributes.URL)
	case "build":
		if event.BuildStatus == "failed" {
			return fmt.Sprintf("🚀 Job status for %s - %s : %s ❌\n%s",
				event.Project.Name, event.BuildName, event.BuildStatus, event.ObjectAttributes.URL)
		}
		if event.BuildStatus == "success" {
			return fmt.Sprintf("🚀 Job status for %s - %s : %s ✅\n%s",
				event.Project.Name, event.BuildName, event.BuildStatus, event.ObjectAttributes.URL)
		}
	default:
		// Unknown type, do not send
		return ""
	}
	return ""
}

func (g *GitlabTelegram) sendTelegramMessage(message string) {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", g.BotToken)
	data := url.Values{
		"chat_id":                  {g.ChatID},
		"text":                     {message},
		"disable_web_page_preview": {"false"},
	}

	resp, err := http.PostForm(apiURL, data)
	if err != nil {
		log.Printf("Error sending message to Telegram: %s", err)
		return
	}
	defer resp.Body.Close()
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	g := GitlabTelegram{
		BotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		ChatID:   os.Getenv("TELEGRAM_CHAT_ID"),
	}
	http.HandleFunc("/", g.HandleWebhook)
	log.Println("Listening at " + os.Getenv("LISTEN_PORT"))
	log.Fatal(http.ListenAndServe(":"+os.Getenv("LISTEN_PORT"), nil))
}
