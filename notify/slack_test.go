package notify

import (
	"fmt"
	"io"
	"net/http"
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
		message string
	}{
		{"chan", "https://localhost/", "", 0, "smile message to slack"},
		{"chan", "https://localhost/", "", 1, "rage message to slack"},
		{"bot", "http://localhost:8080/", "http://localhost/icon.png", 0, "message to slack"},
	}

	for _, s := range expecteds {
		inst := NewSlack(s.channel, s.url, s.iconurl)
		expectPayload := fmt.Sprintf(`{"text":"%s"}`, s.message)

		inst.PostJSON = func(url string, contentType string, body io.Reader) (*http.Response, error) {
			if url != s.url {
				t.Errorf("(expected) '%s' != '%s'", s.url, url)
			}

			if contentType != "application/json" {
				t.Errorf("(expected) 'application/json' != '%s'", contentType)
			}

			bodyBytes, _ := io.ReadAll(body)
			if string(bodyBytes) != expectPayload {
				t.Errorf("(expected) '%s' != '%s'", expectPayload, string(bodyBytes))
			}

			return nil, nil
		}

		err := inst.NotifySlack(s.message, s.status)
		if err != nil {
			t.Error(err)
		}
	}
}
