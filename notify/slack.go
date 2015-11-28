package notify

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type Slack struct {
	Channel    string
	WebhookUrl string
	IconUrl    string
	PostForm   func(string, url.Values) (*http.Response, error)
}

// Slack用のペイロード構造体
type SlackPayload struct {
	Channel string `json:"channel"`
	Text    string `json:"text"`
	Icon    string `json:"icon_url"`
	Emoji   string `json:"icon_emoji"`
}

// コンストラクタ
func NewSlack(channel string, webhookUrl string, iconUrl string) *Slack {
	inst := &Slack{
		Channel:    channel,
		WebhookUrl: webhookUrl,
		IconUrl:    iconUrl,
		PostForm:   http.PostForm,
	}

	return inst
}

// Slackに通知します。
func (self *Slack) NotifySlack(message string, status int) error {
	payload := &SlackPayload{self.Channel, message, "", ""}

	if self.IconUrl != "" {
		payload.Icon = self.IconUrl
	} else {
		if 0 == status {
			payload.Emoji = ":simple_smile:"
		} else {
			payload.Emoji = ":rage:"
		}
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	data := url.Values{}
	data.Add("payload", string(jsonData))

	// Webhool URLにPOSTする
	_, err = self.PostForm(self.WebhookUrl, data)
	if err != nil {
		return err
	}

	return nil
}
