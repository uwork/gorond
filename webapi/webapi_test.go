package webapi

import (
	"bytes"
	"github.com/uwork/gorond/config"
	"github.com/uwork/gorond/goron"
	"github.com/uwork/gorond/logging"
	"github.com/uwork/gorond/util"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	trans, ok := http.DefaultTransport.(*http.Transport)
	if ok {
		trans.DisableKeepAlives = true
	}
	// stubに差し替える
	goron.SystemCommand = func(cmd string, arg ...string) ([]byte, int, error) {
		return []byte(""), 0, nil
	}
}

func TestNewWebApiServer(t *testing.T) {
	addr := ":16777"
	grn := goron.Goron{}
	grn.Config = &config.Config{}

	server, err := NewWebApiServer(addr, &grn)
	if err != nil {
		t.Error(err)
	}

	if server.context != &grn {
		t.Errorf("(expected) server.context(%p) != grn(%p)", server.context, &grn)
	}
	if server.listenAddr != addr {
		t.Errorf("(expected) addr (%s) != (%s)", server.listenAddr, addr)
	}

}

func TestStart(t *testing.T) {
	buffer := &bytes.Buffer{}
	logger = logging.NewLoggerWithWriter(buffer, logging.DEBUG)

	server, err := NewWebApiServer(":16777", nil)
	if err != nil {
		t.Error(err)
		return
	}
	wc := make(chan os.Signal)
	wsc := make(chan error)
	signal.Notify(wc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	err = server.Start(wc, wsc)
	if err != nil {
		t.Error(err)
		return
	}

	// 自分自身にSIGTERMを投げる
	pid := os.Getpid()
	syscall.Kill(pid, syscall.SIGTERM)
	err = <-wsc
	if err.Error() != "accept tcp [::]:16777: use of closed network connection" {
		t.Error(err)
	}

	// terminatedを出力している事を確認
	output := buffer.String()
	if !strings.Contains(output, "terminated") {
		t.Errorf("(exptected) terminated != %s", output)
	}
}

// 404のテスト
func TestResponse404(t *testing.T) {
	server, wc, wsc := createTestServer(t)
	err := server.Start(wc, wsc)
	if err != nil {
		t.Error(err)
	}

	resp, err := http.Get("http://localhost:16777/invalid_path")
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 404 {
		t.Errorf("(expected) status: %d != %d", resp.StatusCode, 404)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	if string(body) != "" {
		t.Error(body)
	}

	pid := os.Getpid()
	syscall.Kill(pid, syscall.SIGTERM)
	err = <-wsc
}

// Jobs APIのテスト
func TestResponseJobs(t *testing.T) {
	server, wc, wsc := createTestServer(t)
	err := server.Start(wc, wsc)
	if err != nil {
		t.Error(err)
	}

	resp, err := http.Get("http://localhost:16777/jobs")
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("(expected) status: %d != %d", resp.StatusCode, 200)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	expected := `{"webapi_test.conf":["* * * * * ? root /bin/echo hello","          - root sleep 3"]}`
	if string(body) != expected {
		t.Errorf("\n\t(expected)%s\n\t(actual)  %s", expected, string(body))
	}
	resp.Body.Close()

	pid := os.Getpid()
	syscall.Kill(pid, syscall.SIGTERM)
	err = <-wsc
}

// Statuses APIのテスト
func TestResponseStatuses(t *testing.T) {
	// sleepを実行するstubに差し替える
	goron.SystemCommand = func(cmd string, args ...string) ([]byte, int, error) {
		if util.ContainsStr("sleep 3", args) {
			time.Sleep(time.Second * 3)
		}
		return []byte(""), 0, nil
	}

	server, wc, wsc := createTestServer(t)
	err := server.Start(wc, wsc)
	if err != nil {
		t.Error(err)
	}

	resp, err := http.Get("http://localhost:16777/statuses")
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("(expected) status: %d != %d", resp.StatusCode, 200)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	expected := `{"webapi_test.conf":{"root /bin/echo hello":"waiting","root sleep 3":"waiting"}}`
	if string(body) != expected {
		t.Errorf("\n\t(expected)%s\n\t(actual)  %s", expected, string(body))
	}
	resp.Body.Close()

	// cronをスタート
	server.context.Start()

	time.Sleep(time.Second)

	// 実行中の状態を取得
	resp2, err := http.Get("http://localhost:16777/statuses")
	if err != nil {
		t.Error(err)
	}
	if resp2.StatusCode != 200 {
		t.Errorf("(expected) status: %d != %d", resp2.StatusCode, 200)
	}
	body2, err := ioutil.ReadAll(resp2.Body)
	if err != nil {
		t.Error(err)
	}
	expected = `{"webapi_test.conf":{"root /bin/echo hello":"waiting","root sleep 3":"running"}}`
	if string(body2) != expected {
		t.Errorf("\n\t(expected)%s\n\t(actual)  %s", expected, string(body2))
	}
	resp2.Body.Close()

	server.context.Stop()

	pid := os.Getpid()
	syscall.Kill(pid, syscall.SIGTERM)
	err = <-wsc
}

// テスト用のサーバを作成
func createTestServer(t *testing.T) (*WebApiServer, chan os.Signal, chan error) {
	conf, err := config.LoadConfig("webapi_test.conf", "config_test.d")
	if err != nil {
		t.Error(err)
	}
	grn, err := goron.NewGorond(conf)
	if err != nil {
		t.Error(err)
	}

	server, err := NewWebApiServer(grn.Config.Config.WebApi, grn)
	if err != nil {
		t.Error(err)
		os.Exit(-1)
	}

	wc := make(chan os.Signal)
	wsc := make(chan error)
	signal.Notify(wc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	return server, wc, wsc
}
