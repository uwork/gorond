package main

import (
	"bytes"
	"github.com/uwork/gorond/goron"
	"github.com/uwork/gorond/logging"
	"github.com/uwork/gorond/util"
	"github.com/uwork/gorond/webapi"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestDoMain(t *testing.T) {

	configPath := "test_config.conf"
	configDir := "test_config.d"
	pidfile := "test.pid"

	configContent := `
[config]
log = ""
cronlog = ""
webapi = 0.0.0.0:6777
notifytype = stdout
notifywhen = onerror
subject = notify result

`

	err := ioutil.WriteFile(configPath, []byte(configContent), 0666)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(configPath)

	err = os.Mkdir(configDir, 0777)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(configDir)

	// 別スレッドで実行
	go doMain(configPath, configDir, pidfile)

	// 起動完了を待つ
	time.Sleep(time.Second)

	// ロガーを横取り
	logger := &bytes.Buffer{}
	cronlogger := &bytes.Buffer{}
	apilogger := &bytes.Buffer{}
	goron.SetLoggerWithWriter(logger, cronlogger, logging.DEBUG)
	webapi.SetLoggerWithWriter(apilogger, logging.DEBUG)

	// pidファイルの存在確認
	if !util.ExistsFile(pidfile) {
		t.Errorf("pid file(%s) not found.", pidfile)
	}

	// 自分自身にSIGTERMを投げる
	pid := os.Getpid()
	syscall.Kill(pid, syscall.SIGTERM)

	// 1秒待つ
	time.Sleep(time.Second)

	// signal: terminated が出力される事を確認
	output := logger.String()
	if !strings.Contains(output, "signal: terminated") {
		t.Errorf("(expected) %v != %v", "signal: terminated", output)
	}

	// apiログにもterminatedが出力される事を確認
	apiout := apilogger.String()
	if !strings.Contains(apiout, "terminated") {
		t.Errorf("(expected) %v != %v", "terminated", apiout)
	}

	// pidファイルの削除確認
	if util.ExistsFile(pidfile) {
		t.Errorf("pid file(%s) not deleted.", pidfile)
	}
}

func TestDoConfigTestSucessful(t *testing.T) {
	configPath := "test_config.conf"
	configDir := "test_config.d"

	configContent := `
[config]
log = ""
cronlog = ""
webapi = 0.0.0.0:6777
notifytype = stdout
notifywhen = onerror
subject = notify result

`
	err := ioutil.WriteFile(configPath, []byte(configContent), 0666)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(configPath)

	err = os.Mkdir(configDir, 0777)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(configDir)

	result := doConfigTest(configPath, configDir)
	if result != 0 {
		t.Error("configtest failure.")
	}

}

func TestDoConfigTestFail1(t *testing.T) {
	configPath := "test_config.conf"
	configDir := "test_config.d"

	configContent := `
log = ""
cronlog = ""
webapi = 0.0.0.0:6777
notifytype = stdout
notifywhen = onerror
subject = notify result

`
	err := ioutil.WriteFile(configPath, []byte(configContent), 0666)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(configPath)

	err = os.Mkdir(configDir, 0777)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(configDir)

	result := doConfigTest(configPath, configDir)
	if result == 0 {
		t.Error("configtest successful.")
	}

}

func TestDoConfigTestFail2(t *testing.T) {
	configPath := "test_config.conf"
	configDir := "test_config.d"

	configContent := `
[config]
log = ""
cronlog = ""
webapi = 0.0.0.0:6777
notifytype = stdout
notifywhen = onerror
subject = notify result

[job]
0 * * * * * error
`
	err := ioutil.WriteFile(configPath, []byte(configContent), 0666)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(configPath)

	err = os.Mkdir(configDir, 0777)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(configDir)

	result := doConfigTest(configPath, configDir)
	if result == 0 {
		t.Error("configtest successful.")
	}

}
