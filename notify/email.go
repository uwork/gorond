package notify

import (
	"fmt"
	"net/smtp"
	"strings"
)

type Mail struct {
	From    string
	To      string
	Subject string
	Body    string
}

type Smtp struct {
	Host      string
	User      string
	Password  string
	PlainAuth func(auth string, user string, password string, host string) smtp.Auth
	SendMail  func(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

const MAIL_FOOTER = `--
This mail sent automatically from gorond.
`

// コンストラクタ
func NewSmtp(host string, user string, pass string) *Smtp {
	instance := &Smtp{
		Host:      host,
		User:      user,
		Password:  pass,
		PlainAuth: smtp.PlainAuth,
		SendMail:  smtp.SendMail,
	}
	return instance
}

// メールを送信する
func (self *Smtp) SendEmail(mail *Mail, status int) error {
	host := strings.Split(self.Host, ":")

	auth := self.PlainAuth(
		"",
		self.User,
		self.Password,
		host[0],
	)

	body := fmt.Sprintf("Subject:%s\r\n\r\n%s\r\n\r\n%s", mail.Subject, mail.Body, MAIL_FOOTER)

	return self.SendMail(
		self.Host,
		auth,
		mail.From,
		strings.Split(mail.To, ","),
		[]byte(body),
	)
}
