package notify

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

// コンストラクタ
func TestNewSlack(t *testing.T) {
	expecteds := []struct {
		channel string
		url     string
		iconurl string
	}{
		{"chan", "https://localhost/", ""},
		{"bot", "http://localhost:8080/", "http://localhost/icon.png"},
	}

	for _, s := range expecteds {
		inst := NewSlack(s.channel, s.url, s.iconurl)
		if inst.Channel != s.channel {
			t.Errorf("(expected) '%s' != '%s'", s.channel, inst.Channel)
		}
		if inst.WebhookUrl != s.url {
			t.Errorf("(expected) '%s' != '%s'", s.url, inst.WebhookUrl)
		}
		if inst.IconUrl != s.iconurl {
			t.Errorf("(expected) '%s' != '%s'", s.iconurl, inst.IconUrl)
		}
	}
}

// Slack通知までのテスト
func TestSendSlack(t *testing.T) {
	expecteds := []struct {
		channel string
		url     string
		iconurl string
		status  int
		emoji   string
		message string
	}{
		{"chan", "https://localhost/", "", 0, ":simple_smile:", "smile message to slack"},
		{"chan", "https://localhost/", "", 1, ":rage:", "rage message to slack"},
		{"bot", "http://localhost:8080/", "http://localhost/icon.png", 0, "", "message to slack"},
	}

	for _, s := range expecteds {
		inst := NewSlack(s.channel, s.url, s.iconurl)
		expectPayload := fmt.Sprintf(`{"channel":"%s","text":"%s","icon_url":"%s","icon_emoji":"%s"}`, s.channel, s.message, s.iconurl, s.emoji)

		inst.PostForm = func(url string, data url.Values) (*http.Response, error) {
			if url != s.url {
				t.Errorf("(expected) '%s' != '%s'", s.url, url)
			}

			payload := data["payload"][0]
			if payload != expectPayload {
				t.Errorf("(expected) '%s' != '%v'", expectPayload, payload)
			}

			return nil, nil
		}

		err := inst.NotifySlack(s.message, s.status)
		if err != nil {
			t.Error(err)
		}
	}
}
