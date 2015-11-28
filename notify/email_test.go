package notify

import (
	"fmt"
	"net/smtp"
	"testing"
)

// コンストラクタのテスト
func TestNewEmail(t *testing.T) {
	expecteds := []struct {
		host string
		user string
		pass string
	}{
		{"test", "user", "pass"},
		{"localhost", "username", "password"},
	}

	for _, s := range expecteds {
		smtp := NewSmtp(s.host, s.user, s.pass)

		if smtp.Host != s.host {
			t.Errorf("(expected) '%s' != '%s'", s.host, smtp.Host)
		}
		if smtp.User != s.user {
			t.Errorf("(expected) '%s' != '%s'", s.user, smtp.User)
		}
		if smtp.Password != s.pass {
			t.Errorf("(expected) '%s' != '%s'", s.pass, smtp.Password)
		}
	}
}

// メール送信直前までのテスト
func TestSendEmail(t *testing.T) {
	sender := NewSmtp("localhost", "user", "pass")
	mail := &Mail{
		"from@localhost",
		"to@localhost",
		"subject message",
		"body message",
	}
	body := fmt.Sprintf("Subject:%s\r\n\r\n%s\r\n\r\n%s", mail.Subject, mail.Body, MAIL_FOOTER)

	sender.PlainAuth = func(auth string, user string, password string, host string) smtp.Auth { return nil }
	sender.SendMail = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {

		if addr != sender.Host {
			t.Errorf("(expected) %s != %s", sender.Host, addr)
		}
		if a != nil {
			t.Errorf("(expected) nil != %v", a)
		}
		if from != mail.From {
			t.Errorf("(expected) %s != %s", mail.From, from)
		}
		if to[0] != mail.To {
			t.Errorf("(expected) %s != %s", mail.To, to[0])
		}
		if string(msg) != body {
			t.Errorf("(expected) %s != %s", body, string(msg))
		}

		return nil
	}

	err := sender.SendEmail(mail, 0)
	if err != nil {
		t.Error(err)
	}
}
