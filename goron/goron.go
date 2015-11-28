package goron

import (
	"github.com/robfig/cron"
	"gorond/config"
	"gorond/fswatch"
	"gorond/logging"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

var logger *logging.Logger
var cronlogger *logging.Logger

type Goron struct {
	cron   *cron.Cron
	Config *config.Config
}

// Gorondを生成します
func NewGorond(config *config.Config) (*Goron, error) {
	goron := &Goron{}
	goron.Config = config

	// set logger
	SetLogger(config.Config.Log, config.Config.CronLog, logging.DEBUG)
	logger.Info("---new gorond---")

	cron, err := newCron(config)
	if err != nil {
		return nil, err
	}
	goron.cron = cron

	return goron, nil
}

// loggerを設定する
func SetLogger(logPath string, cronLogPath string, level logging.Level) {
	logger = logging.NewLogger(logPath, level)
	cronlogger = logging.NewLogger(cronLogPath, level)
}

// loggerを設定する
func SetLoggerWithWriter(logWriter io.Writer, cronLogWriter io.Writer, level logging.Level) {
	logger = logging.NewLoggerWithWriter(logWriter, level)
	cronlogger = logging.NewLoggerWithWriter(cronLogWriter, level)
}

// Cronを開始する
func (self *Goron) Start() {
	logger.Info("## goron start ##")
	self.cron.Start()
}

// Cronを停止する
func (self *Goron) Stop() {
	logger.Info("## goron stop ##")
	self.cron.Stop()
}

// 設定をリロードする
func (self *Goron) Reload(config *config.Config) error {

	cron, err := newCron(config)
	if err != nil {
		return err
	}
	self.Config = config
	self.cron = cron

	// logger reload.
	logger.Close()
	cronlogger.Close()
	SetLogger(config.Config.Log, config.Config.CronLog, logging.DEBUG)

	logger.Infof("## config reloaded ##")

	return nil
}

// configからcronを作成する
func newCron(conf *config.Config) (*cron.Cron, error) {
	cron := cron.New()

	// ジョブを登録する
	for _, job := range conf.Jobs {
		if job.Parent == nil {
			logger.Infof("add job: %s %s %s", job.Schedule, job.User, job.Command)
			err := registerJob(cron, conf, job)
			if err != nil {
				return nil, err
			}
		}
	}

	// 子設定のジョブを登録する
	for _, config := range conf.Childs {
		for _, job := range config.Jobs {
			if job.Parent == nil {
				logger.Infof("add job: %s %s %s", job.Schedule, job.User, job.Command)
				err := registerJob(cron, config, job)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return cron, nil
}

// ジョブをcronに登録する
func registerJob(cron *cron.Cron, conf *config.Config, job *config.Job) error {
	return cron.AddFunc(job.Schedule, func() {
		executionJobs(conf, []*config.Job{job})
	})
}

// ジョブを実行します
func executionJobs(conf *config.Config, jobs []*config.Job) {
	ch := make(chan bool)
	for _, job := range jobs {
		go func(cjob *config.Job) {
			logger.Debugf("execute job: %s", cjob.Command)

			out, status, err := executionCommand(cjob)
			if err != nil {
				Notify(conf, out, status, err, cjob)
			} else if conf.Config.NotifyWhen == "always" {
				Notify(conf, out, status, nil, job)
			}
			ch <- (err == nil)
		}(job)
	}

	// 各ジョブの終了を待つ
	allOk := true
	for _ = range jobs {
		if !<-ch {
			allOk = false
		}
	}

	// 全ジョブが成功した場合、子ジョブを実行する
	if allOk {
		for _, job := range jobs {
			if 0 < len(job.Childs) {
				executionJobs(conf, job.Childs)
			}
		}
	}
}

// コマンドを実行します
func executionCommand(job *config.Job) (string, int, error) {

	// FIXME: ユーザ切り替えの為、常にsuを使ってるけど他に良い方法が無いか検討する
	command := "su"
	os.Getenv("SHELL")
	args := []string{"-", "-c", job.Command}
	if job.User != "" {
		args = []string{"-", job.User, "-c", job.Command}
	}

	job.Status = config.RUNNING
	logger.Debugf("command: %s %s", command, args)
	out, err := exec.Command(command, args...).CombinedOutput()

	output := strings.TrimRight(string(out), "\n")
	if err == nil {
		job.Status = config.WAITING
		cronlogger.Debugf("exec: %s\n\toutput: %s", job.Command, output)
		return output, 0, nil
	} else {
		job.Status = config.FAILED
		if e2, ok := err.(*exec.ExitError); ok {
			if s, ok := e2.Sys().(syscall.WaitStatus); ok {
				return output, s.ExitStatus(), err
			}
		}
		return output, -1, err
	}
}

// コンフィグ自動リロードの設定
func StartAutoReload(grn *Goron, configPath string, includeDir string) (*fswatch.Watcher, error) {
	paths := []string{configPath}
	dirs := map[string]string{includeDir: `.+\.conf`}
	watcher, err := fswatch.StartWatcher(paths, dirs, func(event fswatch.Event) {
		logger.Infof("## config changed: %v", event)

		// 設定をリロードする
		grn.Stop()

		config, err := config.LoadConfig(configPath, includeDir)
		if err != nil {
			logger.Fatal(err)
			os.Exit(-1)
		}
		if err = grn.Reload(config); err != nil {
			logger.Fatal(err)
			os.Exit(-1)
		}

		grn.Start()
	})

	if err != nil {
		return nil, err
	}
	return watcher, nil
}

// シグナルを受け取るまで待つ
func WaitSignal(c chan os.Signal, sc chan int) {

	for {
		s := <-c

		switch s {
		case syscall.SIGHUP:
			logger.Infof("signal: %v", s)

		case syscall.SIGTERM:
			// terminateシグナル
			logger.Infof("signal: %v", s)
			sc <- 15

		case syscall.SIGINT:
			// Ctrl-C
			logger.Infof("signal: %v", s)
			sc <- 2

		default:
			// 不明なシグナル
			logger.Errorf("unknown signal: %v", s)
			sc <- -1
		}
	}

}
