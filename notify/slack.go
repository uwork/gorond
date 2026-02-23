package notify

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Slack struct {
	Channel    string
	WebhookUrl string
	IconUrl    string
	PostJSON   func(url string, contentType string, body io.Reader) (*http.Response, error)
}

// Slack用のペイロード構造体
type SlackPayload struct {
	Text string `json:"text"`
}

// コンストラクタ
func NewSlack(channel string, webhookUrl string, iconUrl string) *Slack {
	inst := &Slack{
		Channel:    channel,
		WebhookUrl: webhookUrl,
		IconUrl:    iconUrl,
		PostJSON:   http.Post,
	}

	return inst
}

// Slackに通知します。
func (self *Slack) NotifySlack(message string, status int) error {
	payload := &SlackPayload{Text: message}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Webhook URLにJSON形式でPOSTする (Slack app形式)
	_, err = self.PostJSON(self.WebhookUrl, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return err
	}

	return nil
}
