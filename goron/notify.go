package goron

import (
	"fmt"
	"gorond/config"
	"gorond/notify"
	"strings"
)

// ジョブ実行後の通知を実行します
func Notify(conf *config.Config, output string, status int, err error, job *config.Job) {

	var subject string
	if err == nil {
		subject = strings.Replace(conf.Config.Subject, "@result", "successful", 1)
	} else {
		subject = strings.Replace(conf.Config.Subject, "@result", "failed", 1)
	}

	body := fmt.Sprintf("command: %s\noutput: %s\nerror: %v", job.Command, output, err)

	switch conf.Config.NotifyType {
	case config.STDOUT:
		notifyStdout(conf, subject, body, status)
	case config.MAIL:
		notifyEmail(conf, subject, body, status)
	case config.SLACK:
		notifySlack(conf, subject, body, status)
	case config.FLUENTD:
		notifyFluentd(conf, subject, body, job.Command, status)
	case config.SNS:
		notifySNS(conf, subject, body, status)
	}
}

// 標準出力に通知する
func notifyStdout(conf *config.Config, subject string, body string, status int) {
	notify.SendStdout(subject, body)
}

// メールで通知する
func notifyEmail(conf *config.Config, subject string, body string, status int) {
	mail := &notify.Mail{
		conf.Mail.From,
		conf.Mail.Dest,
		subject,
		body,
	}

	smtp := notify.NewSmtp(
		conf.Mail.SmtpHost,
		conf.Mail.SmtpUser,
		conf.Mail.SmtpPassword,
	)

	if err := smtp.SendEmail(mail, status); err != nil {
		logger.Error(err)
	}
}

func notifySlack(conf *config.Config, subject string, body string, status int) {
	slack := notify.NewSlack(
		"#"+conf.Slack.Channel,
		conf.Slack.WebhookUrl,
		conf.Slack.IconUrl,
	)

	message := subject + "\n" + body

	if err := slack.NotifySlack(message, status); err != nil {
		logger.Error(err)
	}
}

func notifyFluentd(conf *config.Config, subject string, body string, command string, status int) {
	fluentd := notify.NewFluentd(
		conf.Fluentd.Url,
	)

	if err := fluentd.NotifyFluentd(subject, body, status); err != nil {
		logger.Error(err)
	}
}

func notifySNS(conf *config.Config, subject string, body string, status int) {
	sns := notify.NewSNS(
		conf.SNS.Region,
		conf.SNS.TopicArn,
	)

	if err := sns.NotifySNS(subject, body, status); err != nil {
		logger.Error(err)
	}
}
