package config

import (
	"errors"
	"github.com/uwork/gorond/util"
	"gopkg.in/gcfg.v1"
)

// NotifyType
const (
	STDOUT  = "stdout"
	MAIL    = "mail"
	SLACK   = "slack"
	FLUENTD = "fluentd"
	SNS     = "sns"
)

var NotifyTypes = []string{STDOUT, MAIL, SLACK, FLUENTD, SNS}

// NotifyWhen
const (
	ONERROR = "onerror"
	ALWAYS  = "always"
)

var NotifyWhens = []string{ONERROR, ALWAYS}

// confファイルをパースする
func parseConfig(str string, baseConfig *Config) (*Config, error) {
	config := *baseConfig // 設定をコピー

	err := gcfg.ReadStringInto(&config, str)
	if err != nil {
		return nil, err
	}

	// 列挙値のチェック
	notifyType := config.Config.NotifyType
	if !util.ContainsStr(notifyType, NotifyTypes) {
		return nil, errors.New("invalid NotifyType:" + notifyType)
	}

	notifyWhen := config.Config.NotifyWhen
	if !util.ContainsStr(notifyWhen, NotifyWhens) {
		return nil, errors.New("invalid NotifyWhen:" + notifyWhen)
	}

	return &config, nil
}
