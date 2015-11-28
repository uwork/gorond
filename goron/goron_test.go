package goron

import (
	"bufio"
	"github.com/robfig/cron"
	"gorond/config"
	"gorond/fswatch"
	"gorond/logging"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	testWriter := bufio.NewWriter(ioutil.Discard)
	logger = logging.NewLoggerWithWriter(testWriter, logging.INFO)
	cronlogger = logging.NewLoggerWithWriter(testWriter, logging.INFO)

	// 標準のlogも出力先をDiscardにする
	log.SetOutput(ioutil.Discard)
}

// インスタンス作成のテスト
func TestNewGoron(t *testing.T) {
	conf := createTestConfig()

	goron, err := NewGorond(conf)
	if err != nil {
		t.Errorf("error in NewGorond: %v", err)
	} else if goron == nil {
		t.Error("goron is nil")
	} else if goron.Config != conf {
		t.Error("goron.config:%v != %v", goron.Config, conf)
	}
}

// Cron開始のテスト
func ExampleStart() {
	conf := createTestConfig()
	conf.Jobs = []*config.Job{
		&config.Job{Schedule: "* * * * * *", User: "root", Command: "echo start"},
	}
	goron, _ := NewGorond(conf)
	goron.Start()
	// 1秒待って標準出力をテストする
	time.Sleep(time.Second)
	goron.Stop()
	time.Sleep(time.Second)

	// Output:
	// notify successful: command: echo start
	// output: start
	// error: <nil>
}

// Cron停止のテスト
func ExampleStop() {
	conf := createTestConfig()
	conf.Jobs = []*config.Job{
		&config.Job{Schedule: "* * * * * *", User: "root", Command: "echo stop"},
	}
	goron, _ := NewGorond(conf)
	goron.Start()
	// ジョブ実行を待つ
	time.Sleep(time.Second)
	goron.Stop()
	// さらに3秒間ってジョブ停止を確認する
	time.Sleep(time.Second)

	// Output:
	// notify successful: command: echo stop
	// output: stop
	// error: <nil>
}

// 設定を再読込みさせるテスト
func TestReload(t *testing.T) {
	conf := createTestConfig()
	conf.Jobs = []*config.Job{
		&config.Job{Schedule: "0 * * * * ?", User: "root", Command: "echo reload1"},
	}
	goron, err := NewGorond(conf)

	if err != nil {
		t.Error(err)
	} else if 1 != len(goron.cron.Entries()) {
		t.Error("job register failed.")
	}

	newConf := createTestConfig()
	newConf.Jobs = []*config.Job{
		&config.Job{Schedule: "0 * * * * ?", User: "root", Command: "echo reload2"},
		&config.Job{Schedule: "0 * * * * ?", User: "root", Command: "echo reload3"},
	}

	err = goron.Reload(newConf)

	if err != nil {
		t.Error(err)
	} else if 2 != len(goron.cron.Entries()) {
		t.Error("jobs register failed.")
	}
}

// cronをconfigから生成するテスト
func TestNewCron(t *testing.T) {
	conf := createTestConfig()
	child := createTestConfig()

	conf.Jobs = []*config.Job{
		&config.Job{
			Schedule: "0 * * * * ?",
			User:     "root",
			Command:  "echo newcron1",
		},
	}

	child.Jobs = []*config.Job{
		&config.Job{
			Schedule: "0 * * * * ?",
			User:     "root",
			Command:  "echo newcron2",
		},
	}

	conf.Childs = []*config.Config{child}

	cron, err := newCron(conf)
	if err != nil {
		t.Error(err)
	} else if cron == nil {
		t.Error("cron create failed")
	} else if 0 == len(cron.Entries()) {
		t.Error("cron register job failed.")
	}
}

// ジョブ登録テスト
func TestRegisterJob(t *testing.T) {
	cron := &cron.Cron{}
	conf := createTestConfig()

	job := &config.Job{
		Schedule: "0 * * * * ?",
		User:     "root",
		Command:  "echo registerjob",
	}

	err := registerJob(cron, conf, job)
	if err != nil {
		t.Error(err)
	} else if 0 == len(cron.Entries()) {
		t.Errorf("job entries: %d", len(cron.Entries()))
	}
}

// ジョブ登録失敗テスト
func TestRegisterJobFail(t *testing.T) {
	cron := &cron.Cron{}
	conf := createTestConfig()

	job := &config.Job{
		Schedule: "0 aaa * * * ?",
		User:     "root",
		Command:  "echo registerjob failed",
	}

	err := registerJob(cron, conf, job)

	if err == nil {
		t.Error("register job successful.")
	} else if 0 < len(cron.Entries()) {
		t.Error("register job successful: %d", len(cron.Entries()))
	}
}

// ジョブ実行のテスト
func ExampleExecutionJob() {
	conf := createTestConfig()
	job := config.Job{
		User:    "root",
		Command: "echo test",
	}

	executionJobs(conf, []*config.Job{&job})

	// Output:
	// notify successful: command: echo test
	// output: test
	// error: <nil>
}

// 子ジョブ実行のテスト
func ExampleExecutionJobChilds() {
	conf := createTestConfig()
	job := config.Job{
		User:    "root",
		Command: "echo execjob1",
		Childs: []*config.Job{
			&config.Job{
				User:    "root",
				Command: "echo execjob2",
			},
		},
	}

	executionJobs(conf, []*config.Job{&job})

	// Output:
	// notify successful: command: echo execjob1
	// output: execjob1
	// error: <nil>
	// notify successful: command: echo execjob2
	// output: execjob2
	// error: <nil>
}

// ジョブ実行失敗のテスト
func ExampleExecutionJobFailed() {
	conf := createTestConfig()
	job := config.Job{
		User:    "root",
		Command: "grep x execfailtest",
	}

	executionJobs(conf, []*config.Job{&job})

	// Output:
	// notify failed: command: grep x execfailtest
	// output: grep: execfailtest: No such file or directory
	// error: exit status 2
}

// コマンド実行のテスト
func TestExecutionCommand(t *testing.T) {

	job := config.Job{
		User:    "root",
		Command: "echo cmdtest",
	}

	out, status, err := executionCommand(&job)

	if err != nil {
		t.Error(err)
	}
	if 0 != status {
		t.Errorf("status: %d (not 0)", status)
	}
	if out != "cmdtest" {
		t.Errorf("output: %s", out)
	}

}

// コマンドエラー時のテスト
func TestExecutionCommandFail(t *testing.T) {

	job := config.Job{
		User:    "root",
		Command: "grep x execcmdfail",
	}

	out, status, err := executionCommand(&job)

	if err == nil {
		t.Error(err)
	}
	if 2 != status {
		t.Errorf("status: %d (not 2)", status)
	}
	if out != "grep: execcmdfail: No such file or directory" {
		t.Errorf("output: %s", out)
	}

}

// 設定自動リロードのテスト
func TestStartAutoReload(t *testing.T) {
	configPath := ".reload.conf"
	configDir := "reload.d"
	os.Mkdir(configDir, 0777)
	ioutil.WriteFile(configPath, []byte(`
[config]
notifytype = stdout
notifywhen = always
webapi = 0.0.0.0:6777
log = ""
cronlog = ""
`), 0666)
	defer os.Remove(configPath)
	defer os.Remove(configDir)

	conf, err := config.LoadConfig(configPath, configDir)
	if err != nil {
		t.Error(err)
	}
	// 設定の確認
	if conf.Config.NotifyWhen != "always" {
		t.Errorf("config load failed. notifyWhen: %v", conf.Config.NotifyWhen)
	}
	goron, err := NewGorond(conf)
	if err != nil {
		t.Error(err)
	}
	goron.Start()
	defer goron.Stop()

	// 1秒毎に監視
	fswatch.WatchInterval = time.Second
	_, err = StartAutoReload(goron, configPath, configDir)
	if err != nil {
		t.Error(err)
	}

	// 2秒待って設定を書き換える
	time.Sleep(time.Second * 2)
	ioutil.WriteFile(configPath, []byte(`
[config]
notifytype = stdout
notifywhen = onerror
webapi = 0.0.0.0:6777
log = ""
cronlog = ""
`), 0666)

	// さらに2秒待つ
	time.Sleep(time.Second * 2)

	// 書き換わっている事の確認
	if goron.Config.Config.NotifyWhen != "onerror" {
		t.Errorf("config not reloaded. notifyWhen: %v", goron.Config.Config.NotifyWhen)
	}
}

// シグナル待ちのテスト
func TestWaitSignal(t *testing.T) {

	c := make(chan os.Signal)
	sc := make(chan int, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go WaitSignal(c, sc)

	c <- syscall.SIGTERM

	if status := <-sc; 15 != status {
		t.Errorf("invalid status code: %d", status)
	}
}

func createTestConfig() *config.Config {
	conf := config.Config{
		Config: config.SettingConfig{
			NotifyType: config.STDOUT,
			NotifyWhen: "always",
			Subject:    "notify @result",
		},
	}

	return &conf
}
