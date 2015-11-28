package notify

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type Fluentd struct {
	Url      string
	PostForm func(string, url.Values) (*http.Response, error)
}

// Fluentd用のペイロード構造体
type FluentdMessage struct {
	Subject string `json:"subject"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// コンストラクタ
func NewFluentd(url string) *Fluentd {
	inst := &Fluentd{
		Url:      url,
		PostForm: http.PostForm,
	}

	return inst
}

// Fluentdに通知します。
func (self *Fluentd) NotifyFluentd(subject string, message string, status int) error {
	msg := &FluentdMessage{subject, message, status}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	data := url.Values{}
	data.Add("json", string(jsonData))

	// Webhool URLにPOSTする
	_, err = self.PostForm(self.Url, data)
	if err != nil {
		return err
	}

	return nil
}
