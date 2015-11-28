package config

import (
	"errors"
	"github.com/uwork/gorond/util"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type SettingConfig struct {
	Log        string
	CronLog    string
	ApiLog     string
	WebApi     string
	NotifyType string
	NotifyWhen string
	Subject    string
}

type MailConfig struct {
	Dest         string
	From         string
	SmtpHost     string
	SmtpUser     string
	SmtpPassword string
}

type SlackConfig struct {
	Channel    string
	WebhookUrl string
	IconUrl    string
}

type FluentdConfig struct {
	Url string
}

type SNSConfig struct {
	Region   string
	TopicArn string
}

// Goron設定構造体
type Config struct {
	File string

	Config  SettingConfig
	Mail    MailConfig
	Slack   SlackConfig
	Fluentd FluentdConfig
	SNS     SNSConfig

	Jobs   []*Job
	Childs []*Config
}

// 設定を読み込んで返す。
func LoadConfig(configPath string, includeDir string) (*Config, error) {
	root, err := newGoronConfig(configPath, &Config{})
	if err != nil {
		return nil, err
	}

	confFiles, err := util.FileList(includeDir, `.+\.conf`)
	if err != nil {
		return nil, err
	}

	for _, confPath := range confFiles {
		config, err := newGoronConfig(includeDir+"/"+confPath, root)
		if err != nil {
			return nil, err
		}

		root.Childs = append(root.Childs, config)
	}

	return root, nil
}

// 新しい設定ファイルを読み込む
func newGoronConfig(configPath string, baseConfig *Config) (*Config, error) {
	if stat, err := os.Stat(configPath); stat == nil || stat.IsDir() {
		log.Println(err)
		return nil, errors.New(configPath + " is not found")
	}

	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	configString, jobsString, err := parseConfigString(string(content))
	if err != nil {
		return nil, err
	}

	// iniファイル部分をパース
	config, err := parseConfig(configString, baseConfig)
	if err != nil {
		return nil, err
	}

	// job設定部分をパース
	jobs, err := parseJobConfig(jobsString)
	if err != nil {
		return nil, err
	}
	config.Jobs = jobs
	config.File = configPath

	//log.Printf("readed config: %s\n", configPath)

	return config, nil
}

// 設定ファイルを読んで設定とジョブの文字列に分ける
func parseConfigString(config string) (string, string, error) {

	// 設定とジョブを分割して返す.
	if strings.Contains(config, "[job]") {
		contents := strings.Split(config, "[job]")
		return strings.Trim(contents[0], "\n"), strings.Trim(contents[1], "\n"), nil
	} else {
		return strings.Trim(config, "\n"), "", nil
	}
}
