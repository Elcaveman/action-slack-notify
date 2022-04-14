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

func slackDivider() string {
	return `{"type":"divider"}`
}
func slackContext(authorName string,authorIcon string,authorLink string) string {
	return fmt.Sprintf(`{
		"type":"context",
		"elements": [
			{
				"type": "image",
				"image_url": "%s",
				"alt_text": "%s"
			},
			{
				"type": "mrkdwn",
				"text": "*By* <%s|%s>"
			}
		]
	}`,authorIcon,authorName,authorLink,authorName)
}
func slackSection(text string) string {
	return fmt.Sprintf(`{
		"type":"section",
		"text": {
			"type": "mrkdwn",
			"text": "%s"
		}
	}`,text)
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
	
	var scope string
	if (os.Getenv("GITHUB_HEAD_REF")=="main"){
		scope = "🚀 Prod"
	}else{
		scope = "🚧 "+os.Getenv("GITHUB_HEAD_REF")
	}
	color := ""
	switch os.Getenv(EnvSlackColor) {
	case "success":
		color = "good"
	case "cancelled":
		color = "#808080"
	case "failure":
		color = "danger"
	default:
		color = EnvSlackColor
	}
	fmt.Print(color)
	impact := "🔄 Server will reboot"
	msg := fmt.Sprintf(`{
		"blocks":[
			%s ,
			%s ,
			%s ,
			%s ,
			%s ,
			%s ,
			%s
		]}`,
		slackSection("*Action:* Merge hotfixes into *"+os.Getenv(envSolutionName)+"*"),
		slackSection("*Message:* _"+commit_message+"_ "),
		slackSection("*Impact:* " + impact),
		slackSection("*Scope:* " + scope),
		slackSection("*Status:* " + status),
		slackDivider(),
		slackContext(
			os.Getenv(EnvGithubActor),
			os.Getenv("GITHUB_SERVER_URL") + "/" + os.Getenv(EnvGithubActor) + ".png?size=50",
			os.Getenv("GITHUB_SERVER_URL") + "/" + os.Getenv(EnvGithubActor)))
	if err := send(endpoint, msg); err != nil {
		fmt.Fprintf(os.Stderr, "Error sending message: %s\n", err)
		os.Exit(2)
	}
}

func send(endpoint string, msg string) error {
	// enc, err := json.Marshal(msg)
	// if err != nil {
	// 	return err
	// }
	fmt.Print(msg)
	b := bytes.NewBuffer([]byte(msg))
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