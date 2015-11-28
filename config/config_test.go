package config

import (
	"testing"
)

// 設定文字列をパースするテスト
func TestParseConfigString(t *testing.T) {
	expecteds := []struct {
		test      string
		expConfig string
		expJob    string
	}{
		{
			`[config]
log = /var/log/gorond/goron.log

[job]
0 * * * * * root /bin/echo test
`,
			"[config]\nlog = /var/log/gorond/goron.log",
			"0 * * * * * root /bin/echo test",
		},
		{
			`[config]
log = /var/log/gorond/goron.log
cronlog = /var/log/gorond/cron.log

`,
			"[config]\nlog = /var/log/gorond/goron.log\ncronlog = /var/log/gorond/cron.log",
			"",
		},
		{
			`[job]
0 * * * * ? user /bin/echo hello
`,
			"",
			"0 * * * * ? user /bin/echo hello",
		},
	}

	for _, s := range expecteds {
		config, job, err := parseConfigString(s.test)

		if err != nil {
			t.Errorf("error: %v", err)
		}
		if config != s.expConfig {
			t.Errorf("config mismatch (expected) '%s' != '%s'", s.expConfig, config)
		}
		if job != s.expJob {
			t.Errorf("job mismatch (expected) '%s' != '%s'", s.expJob, job)
		}
	}
}

// 設定を読み込むテスト
func TestLoadConfig(t *testing.T) {

	expected := &Config{
		File: "config_test.conf",
		Config: SettingConfig{
			"/var/log/gorond/goron.log",
			"/var/log/gorond/goron_cron.log",
			"/var/log/gorond/goron_api.log",
			"0.0.0.0:6777",
			"mail",
			"onerror",
			"[gorond] job @result",
		},
		Mail: MailConfig{
			"to@example.com",
			"from@example.com",
			"localhost:25",
			"user",
			"password",
		},
		Slack: SlackConfig{
			"bot",
			"https://hooks.slack.com/services/XXXXXXXXXXXXX/XXXXXXXXXXX",
			"",
		},
		Fluentd: FluentdConfig{
			"http://localhost:8888/tag",
		},
		SNS: SNSConfig{
			"ap-northeast-1",
			"arn:aws:sns:ap-northeast-1:9999999999:goron_test",
		},
	}

	jobExpected := *expected
	jobExpected.Config = SettingConfig{
		"/var/log/gorond/goron.log",
		"/var/log/gorond/goron_cron.log",
		"/var/log/gorond/goron_api.log",
		"0.0.0.0:6777",
		"slack",
		"always",
		"[gorond] job @result",
	}
	jobExpected.Jobs = []*Job{
		&Job{Schedule: "0 * * * * ?", User: "vagrant", Command: "/bin/echo test"},
		&Job{Schedule: "10 * * * * ?", User: "root", Command: "echo start",
			Childs: []*Job{
				&Job{Indent: 12, User: "root", Command: "sleep 2; echo ok2"},
				&Job{Indent: 12, User: "root", Command: "sleep 1; echo ok1",
					Childs: []*Job{
						&Job{Indent: 14, User: "root", Command: "echo ok3"},
					},
				},
			},
		},
	}

	config, err := LoadConfig("config_test.conf", "config_test.d")

	if err != nil {
		t.Errorf("error: %s", err)
	}
	if expected.File != config.File {
		t.Errorf("attribute mismatch (expected) '%s' != '%s'", expected.File, config.File)
	}
	if expected.Config != config.Config {
		t.Errorf("attribute mismatch (expected) '%s' != '%s'", expected.Config, config.Config)
	}
	if expected.Mail != config.Mail {
		t.Errorf("attribute mismatch (expected) '%s' != '%s'", expected.Config, config.Config)
	}
	if expected.Slack != config.Slack {
		t.Errorf("attribute mismatch (expected) '%s' != '%s'", expected.Slack, config.Slack)
	}
	if expected.Fluentd != config.Fluentd {
		t.Errorf("attribute mismatch (expected) '%s' != '%s'", expected.Fluentd, config.Fluentd)
	}
	if expected.SNS != config.SNS {
		t.Errorf("attribute mismatch (expected) '%s' != '%s'", expected.SNS, config.SNS)
	}
	if 1 != len(config.Childs) {
		t.Errorf("attribute mismatch (expected) '%v'", config.Childs)
	}

	// job.confのテスト
	if jobExpected.Config != config.Childs[0].Config {
		t.Errorf("attribute mismatch (expected) '%v' != '%v'", jobExpected.Config, config.Childs[0].Config)
	}
	if jobExpected.Mail != config.Childs[0].Mail {
		t.Errorf("attribute mismatch (expected) '%v' != '%v'", jobExpected.Mail, config.Childs[0].Mail)
	}

	// Job1層のテスト
	if !jobEqual(jobExpected.Jobs[0], config.Childs[0].Jobs[0]) {
		t.Errorf("jobs mismatch (expected) '%v' != '%v'", jobExpected.Jobs[0], config.Childs[0].Jobs[0])
	}
	if !jobEqual(jobExpected.Jobs[1], config.Childs[0].Jobs[1]) {
		t.Errorf("jobs mismatch (expected) '%v' != '%v'", jobExpected.Jobs[1], config.Childs[0].Jobs[1])
	}
	// Job2層のテスト
	jobExpected2 := jobExpected.Jobs[1].Childs
	configJobs2 := config.Childs[0].Jobs[1].Childs
	if !jobEqual(jobExpected2[0], configJobs2[0]) {
		t.Errorf("jobs mismatch (expected) '%v' != '%v'", jobExpected2[0], configJobs2[0])
	}
	if !jobEqual(jobExpected2[1], configJobs2[1]) {
		t.Errorf("jobs mismatch (expected) '%v' != '%v'", jobExpected2[1], configJobs2[1])
	}

	// Job3層のテスト
	jobExpected3 := jobExpected2[1].Childs
	configJobs3 := configJobs2[1].Childs
	if !jobEqual(jobExpected3[0], configJobs3[0]) {
		t.Errorf("jobs mismatch (expected) '%v' != '%v'", jobExpected3[0], configJobs3[0])
	}
}

func jobEqual(left *Job, right *Job) bool {
	if left.Schedule != right.Schedule {
		return false
	}
	if left.User != right.User {
		return false
	}
	if left.Command != right.Command {
		return false
	}
	if left.Indent != right.Indent {
		return false
	}
	return true
}
