package notify

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

// コンストラクタ
func TestNewFluentd(t *testing.T) {
	expecteds := []struct {
		url string
	}{
		{"https://localhost/"},
		{"http://localhost:8080/"},
	}

	for _, s := range expecteds {
		inst := NewFluentd(s.url)
		if inst.Url != s.url {
			t.Errorf("(expected) '%s' != '%s'", s.url, inst.Url)
		}
	}
}

// 通知までのテスト
func TestNotifyFluentd(t *testing.T) {
	expecteds := []struct {
		url     string
		subject string
		message string
		status  int
	}{
		{"http://localhost/", "subject", "message", 0},
		{"https://localhost/", "サブジェクト", "メッセージ", 1},
		{"http://localhost:8080/", "test successful", "command: echo test", 0},
	}

	for _, s := range expecteds {
		inst := NewFluentd(s.url)
		expectPayload := fmt.Sprintf(`{"subject":"%s","message":"%s","status":%d}`, s.subject, s.message, s.status)

		inst.PostForm = func(url string, data url.Values) (*http.Response, error) {
			if url != s.url {
				t.Errorf("(expected) '%s' != '%s'", s.url, url)
			}

			payload := data["json"][0]
			if payload != expectPayload {
				t.Errorf("(expected) '%s' != '%v'", expectPayload, payload)
			}

			return nil, nil
		}

		err := inst.NotifyFluentd(s.subject, s.message, s.status)
		if err != nil {
			t.Error(err)
		}
	}
}
