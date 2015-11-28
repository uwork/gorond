package logging

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

// ログレベルのテスト
func TestLogLevel(t *testing.T) {
	testLogFile := "goron_test.log"
	os.Remove(testLogFile)
	logger := NewLogger(testLogFile, INFO)

	logger.Debug("debug log")
	logger.Info("info log")
	logger.Error("error log")
	logger.Fatal("fatal log")

	logs, _ := ioutil.ReadFile(testLogFile)
	logstr := string(logs)

	if strings.Contains(logstr, "debug log") {
		t.Errorf("debug log found:\n\n%s", logstr)
	}
	if !strings.Contains(logstr, "info log") {
		t.Errorf("info log not found:\n\n%s", logstr)
	}
	if !strings.Contains(logstr, "error log") {
		t.Errorf("error log not found:\n\n%s", logstr)
	}
	if !strings.Contains(logstr, "fatal log") {
		t.Errorf("fatal log not found:\n\n%s", logstr)
	}

	os.Remove(testLogFile)
}

// Writer指定のテスト
func TestNewLoggerWithWriter(t *testing.T) {
	buf := &bytes.Buffer{}

	logger := NewLoggerWithWriter(buf, INFO)

	logger.Debug("debug log")
	logger.Info("info log")
	logger.Error("error log")
	logger.Fatal("fatal log")

	logstr := buf.String()

	if strings.Contains(logstr, "debug log") {
		t.Errorf("debug log found:\n\n%s", logstr)
	}
	if !strings.Contains(logstr, "info log") {
		t.Errorf("info log not found:\n\n%s", logstr)
	}
	if !strings.Contains(logstr, "error log") {
		t.Errorf("error log not found:\n\n%s", logstr)
	}
	if !strings.Contains(logstr, "fatal log") {
		t.Errorf("fatal log not found:\n\n%s", logstr)
	}
}
