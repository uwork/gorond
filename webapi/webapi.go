package webapi

import (
	"encoding/json"
	"fmt"
	"github.com/uwork/gorond/config"
	"github.com/uwork/gorond/goron"
	"github.com/uwork/gorond/logging"
	"io"
	"net"
	"net/http"
	"os"
	"syscall"
)

type WebApiServer struct {
	listenAddr string
	server     *http.ServeMux
	context    *goron.Goron
	listener   *net.Listener
}

type JobsResponse map[string][]string

type StatusesResponse map[string]map[string]string

var logger = logging.NewLogger("", logging.DEBUG)

func NewWebApiServer(addr string, context *goron.Goron) (*WebApiServer, error) {
	server := &WebApiServer{
		listenAddr: addr,
		server:     http.NewServeMux(),
		context:    context,
	}

	return server, nil
}

// ログを書き換える
func SetLogger(logPath string) {
	if logger != nil {
		logger.Close()
	}
	logger = logging.NewLogger(logPath, logging.DEBUG)
}

func SetLoggerWithWriter(writer io.Writer, level logging.Level) {
	if logger != nil {
		logger.Close()
	}
	logger = logging.NewLoggerWithWriter(writer, level)
}

// WebApiサーバを開始します
func (self *WebApiServer) Start(sig chan os.Signal, ch chan error) error {
	listener, err := net.Listen("tcp", self.listenAddr)
	if err != nil {
		return err
	}
	self.listener = &listener

	self.server.HandleFunc("/statuses", self.handleStatuses)
	self.server.HandleFunc("/jobs", self.handleJobs)
	self.server.HandleFunc("/", self.response404)

	// 停止時のエラー内容をチャネルに送信
	go func() {
		err := http.Serve(listener, self.server)
		if err != nil {
			logger.Error(err)
		}
		ch <- err
	}()

	// SIGTERMシグナルをハンドリングしてサーバーを終了させる
	go func() {
		for {
			s := <-sig
			switch s {
			case syscall.SIGTERM:
				fallthrough
			case syscall.SIGINT:
				logger.Error(s)
				listener.Close()
			default:
				logger.Error(s)
			}
		}
	}()

	return nil
}

func (self *WebApiServer) response404(w http.ResponseWriter, req *http.Request) {
	logger.Errorf("not found: %s", req.RequestURI)
	w.WriteHeader(404)
	w.Write([]byte(""))
}

func (self *WebApiServer) response500(w http.ResponseWriter, req *http.Request, err error) {
	logger.Infof("access: %s", req.RequestURI)
	w.WriteHeader(500)
	w.Write([]byte(""))
}

func (self *WebApiServer) handleJobs(w http.ResponseWriter, req *http.Request) {
	logger.Infof("access: %s", req.RequestURI)

	resp := JobsResponse{}

	items := createJobsResponseItems(self.context.Config)
	resp[self.context.Config.File] = items

	for _, conf := range self.context.Config.Childs {
		_items := createJobsResponseItems(conf)
		resp[conf.File] = _items
	}

	json, err := json.Marshal(resp)
	if err != nil {
		self.response500(w, req, err)
	} else {
		w.WriteHeader(200)
		w.Write(json)
	}
}

func (self *WebApiServer) handleStatuses(w http.ResponseWriter, req *http.Request) {
	logger.Infof("access: %s", req.RequestURI)

	resp := StatusesResponse{}

	items := createStatusesResponseItems(self.context.Config)
	resp[self.context.Config.File] = items

	for _, conf := range self.context.Config.Childs {
		_items := createStatusesResponseItems(conf)
		resp[conf.File] = _items
	}

	json, err := json.Marshal(resp)
	if err != nil {
		self.response500(w, req, err)
	} else {
		w.WriteHeader(200)
		w.Write(json)
	}
}

func createJobsResponseItems(conf *config.Config) []string {
	jobs := []string{}
	for _, job := range conf.Jobs {
		jobs = append(jobs, job.Line)
	}
	return jobs
}

func createStatusesResponseItems(conf *config.Config) map[string]string {
	items := map[string]string{}
	for _, job := range conf.Jobs {
		key := fmt.Sprintf("%s %s", job.User, job.Command)
		items[key] = job.Status
	}
	return items
}
