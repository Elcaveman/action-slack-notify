package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const (
	EnvSlackWebhook   = "SLACK_WEBHOOK"
	EnvSlackIcon      = "SLACK_ICON"
	EnvSlackIconEmoji = "SLACK_ICON_EMOJI"
	EnvSlackChannel   = "SLACK_CHANNEL"
	// EnvSlackTitle     = "SLACK_TITLE"
	// EnvSlackMessage   = "SLACK_MESSAGE"
	EnvSlackColor     = "SLACK_COLOR"
	EnvSlackUserName  = "SLACK_USERNAME"
	EnvSlackFooter    = "SLACK_FOOTER"
	EnvGithubActor    = "GITHUB_ACTOR"
	EnvSiteName       = "SITE_NAME"
	EnvHostName       = "HOST_NAME"
	EnvMinimal        = "MSG_MINIMAL"
	EnvSlackLinkNames = "SLACK_LINK_NAMES"
	EnvGithubHeadCommitMessage = "GIT_HEAD_COMMIT_MESSAGE"
	EnvJobStatus = "STATUS"
	envSolutionName = "SOLUTION"
)

type Webhook struct {
	Text        string       `json:"text,omitempty"`
	UserName    string       `json:"username,omitempty"`
	IconURL     string       `json:"icon_url,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Channel     string       `json:"channel,omitempty"`
	LinkNames   string       `json:"link_names,omitempty"`
	UnfurlLinks bool         `json:"unfurl_links"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Fallback   string  `json:"fallback"`
	Pretext    string  `json:"pretext,omitempty"`
	Color      string  `json:"color,omitempty"`
	AuthorName string  `json:"author_name,omitempty"`
	AuthorLink string  `json:"author_link,omitempty"`
	AuthorIcon string  `json:"author_icon,omitempty"`
	Footer     string  `json:"footer,omitempty"`
	Fields     []Field `json:"fields,omitempty"`
}

type Field struct {
	Title string `json:"title,omitempty"`
	Value string `json:"value,omitempty"`
	Short bool   `json:"short,omitempty"`
}
func main(){
	endpoint := os.Getenv(EnvSlackWebhook)
	if endpoint == "" {
		fmt.Fprintln(os.Stderr, "URL is required")
		os.Exit(1)
	}
	if strings.HasPrefix(os.Getenv("GITHUB_WORKFLOW"), ".github") {
		os.Setenv("GITHUB_WORKFLOW", "Link to action run")
	}
	long_sha := os.Getenv("GITHUB_SHA")
	commit_message := os.Getenv(EnvGithubHeadCommitMessage)
	var status string
	if (os.Getenv(EnvJobStatus)=="success"){
		status = ":coche-verte: @here"
	}else if (os.Getenv(EnvJobStatus)=="failure"){
		status = ":angryllaume: You're Fired! @" + os.Getenv(EnvGithubActor)
	}else if (os.Getenv(EnvJobStatus)=="cancelled"){
		status = ":warning: Job Cancelled By @" + os.Getenv(EnvGithubActor)
	}else{
		status = ":loading:"
	}
	
	var head_ref string
	if (os.Getenv("GITHUB_HEAD_REF")=="main"){
		head_ref = "🚀 Prod - " + os.Getenv(envSolutionName) 
	}else{
		head_ref = "🚧 "+os.Getenv("GITHUB_HEAD_REF")+ " - "+ os.Getenv(envSolutionName) 
	}
	fields := []Field{}
	mainFields := []Field{
		{
			Title: "Scope",
			Value: head_ref,
			Short: false,
		},
		{
			Title: "Commit",
			Value: "<" + os.Getenv("GITHUB_SERVER_URL") + "/" + os.Getenv("GITHUB_REPOSITORY") + "/commit/" + long_sha + "|" + commit_message + ">",
			Short: false,
		},
		{
			Title: "Impact",
			Value: "Server will restart smoothly",
			Short: false,
		},
		{
			Title: "Status",
			Value: status,
			Short: false,
		},
	}
	fields = append(mainFields, fields...)
	
	color := ""
	switch os.Getenv(EnvSlackColor) {
	case "success":
		color = "good"
	case "cancelled":
		color = "#808080"
	case "failure":
		color = "danger"
	default:
		color = envOr(EnvSlackColor, "good")
	}

	msg := Webhook{
		UserName:  os.Getenv(EnvSlackUserName),
		IconURL:   os.Getenv(EnvSlackIcon),
		IconEmoji: os.Getenv(EnvSlackIconEmoji),
		Channel:   os.Getenv(EnvSlackChannel),
		LinkNames: os.Getenv(EnvSlackLinkNames),
		Attachments: []Attachment{
			{
				Color:      color,
				AuthorName: envOr(EnvGithubActor, ""),
				AuthorLink: os.Getenv("GITHUB_SERVER_URL") + "/" + os.Getenv(EnvGithubActor),
				AuthorIcon: os.Getenv("GITHUB_SERVER_URL") + "/" + os.Getenv(EnvGithubActor) + ".png?size=32",
				Footer:     envOr(EnvSlackFooter, "<https://github.com/rtCamp/github-actions-library|Powered By rtCamp's GitHub Actions Library>"),
				Fields:     fields,
			},
		},
	}

	if err := send(endpoint, msg); err != nil {
		fmt.Fprintf(os.Stderr, "Error sending message: %s\n", err)
		os.Exit(2)
	}
}

func envOr(name, def string) string {
	if d, ok := os.LookupEnv(name); ok {
		return d
	}
	return def
}

func send(endpoint string, msg Webhook) error {
	enc, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	b := bytes.NewBuffer(enc)
	res, err := http.Post(endpoint, "application/json", b)
	if err != nil {
		return err
	}

	if res.StatusCode >= 299 {
		return fmt.Errorf("Error on message: %s\n", res.Status)
	}
	fmt.Println(res.Status)
	return nil
}
