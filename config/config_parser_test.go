package config

import (
	"testing"
)

func TestParseConfig(t *testing.T) {

	test := `
[config]
webapi = 0.0.0.0:6777

log = /var/log/gorond/goron.log
cronlog = /var/log/gorond/cron.log
apilog = /var/log/gorond/api.log

notifytype = mail
notifywhen = always

subject = "[gorond] job test: @result"

[mail]
dest = to@example.com
from = from@example.com
smtphost = localhost:25
smtpuser = user
smtppassword = password

[slack]
channel = test
webhookurl = https://hooks.slack.com/services/XXXXX/XXXXXXX/XXXXXXXXXXXXXXXXX
iconurl = http://example.com/icon.png

[fluentd]
url = http://localhost:8888/tag

[sns]
region = ap-northeast-1
topicarn = arn:aws:sns:ap-northeast-1:9999999999999:test_topic

`

	exp := &Config{
		"",
		SettingConfig{
			"/var/log/gorond/goron.log",
			"/var/log/gorond/cron.log",
			"/var/log/gorond/api.log",
			"0.0.0.0:6777",
			"mail",
			"always",
			"[gorond] job test: @result",
		},
		MailConfig{
			"to@example.com",
			"from@example.com",
			"localhost:25",
			"user",
			"password",
		},
		SlackConfig{
			"test",
			"https://hooks.slack.com/services/XXXXX/XXXXXXX/XXXXXXXXXXXXXXXXX",
			"http://example.com/icon.png",
		},
		FluentdConfig{
			"http://localhost:8888/tag",
		},
		SNSConfig{
			"ap-northeast-1",
			"arn:aws:sns:ap-northeast-1:9999999999999:test_topic",
		},
		[]*Job{},
		[]*Config{},
	}

	config, _ := parseConfig(test, &Config{})
	if config.File != exp.File {
		t.Errorf("mismatch (expected) %s != %s", exp.File, config.File)
	}
	if config.Config != exp.Config {
		t.Errorf("mismatch (expected) %s != %s", exp.Config, config.Config)
	}
	if config.Mail != exp.Mail {
		t.Errorf("mismatch (expected) %s != %s", exp.Mail, config.Mail)
	}
	if config.Slack != exp.Slack {
		t.Errorf("mismatch (expected) %s != %s", exp.Slack, config.Slack)
	}
	if config.Fluentd != exp.Fluentd {
		t.Errorf("mismatch (expected) %s != %s", exp.Fluentd, config.Fluentd)
	}
	if config.SNS != exp.SNS {
		t.Errorf("mismatch (expected) %s != %s", exp.SNS, config.SNS)
	}
}
