package logging

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
)

type Level uint8

const (
	_ Level = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
	NONE
)

type Logger struct {
	level  Level
	logger *log.Logger
	writer *bufio.Writer
	file   *os.File
}

// 新規ロガーの作成
func NewLogger(path string, level Level) *Logger {
	if path != "" {
		// ファイルに出力する
		file := openLogFile(path)
		writer := bufio.NewWriter(file)
		logger := log.New(writer, "", log.LstdFlags)

		return &Logger{level, logger, writer, file}
	} else {
		// devnullに出力する
		logger := log.New(ioutil.Discard, "", log.LstdFlags)
		return &Logger{level, logger, nil, nil}
	}
}

// 新規ロガーの作成
func NewLoggerWithWriter(writer io.Writer, level Level) *Logger {
	bufwriter := bufio.NewWriter(writer)
	newlog := log.New(bufwriter, "", log.LstdFlags)

	logger := &Logger{level, newlog, bufwriter, nil}
	return logger
}

// ロガーをクローズする
func (self *Logger) Close() {
	if self.file != nil {
		self.file.Close()
	}
}

// DEBUGレベルのログを出力する
func (self *Logger) Debug(v ...interface{}) {
	self.log(DEBUG, v...)
}

// DEBUGレベルのログを出力する
func (self *Logger) Debugf(format string, args ...interface{}) {
	self.logf(DEBUG, format, args...)
}

// INFOレベルのログを出力する
func (self *Logger) Info(v ...interface{}) {
	self.log(INFO, v...)
}

// INFOレベルのログを出力する
func (self *Logger) Infof(format string, args ...interface{}) {
	self.logf(INFO, format, args...)
}

// ERRORレベルのログを出力する
func (self *Logger) Error(v ...interface{}) {
	self.log(ERROR, v...)
}

// ERRORレベルのログを出力する
func (self *Logger) Errorf(format string, args ...interface{}) {
	self.logf(ERROR, format, args...)
}

// FATALレベルのログを出力する
func (self *Logger) Fatal(v ...interface{}) {
	self.log(FATAL, v...)
}

// FATALレベルのログを出力する
func (self *Logger) Fatalf(format string, args ...interface{}) {
	self.logf(FATAL, format, args...)
}

// ログファイルをオープンする
func openLogFile(filepath string) *os.File {
	// ディレクトリを作成
	dir := path.Dir(filepath)
	if stat, err := os.Stat(dir); stat != nil {
		if !stat.IsDir() {
			if err = os.Mkdir(dir, 0755); err != nil {
				log.Println(err)
				os.Exit(-1)
			}
		}
	}

	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}

	return f
}

// ログレベルを文字列として取得する
func levelToString(level Level) string {
	switch level {
	case INFO:
		return "[INFO]"
	case DEBUG:
		return "[DEBUG]"
	case ERROR:
		return "[ERROR]"
	case FATAL:
		return "[FATAL]"
	}
	return "[UNKNOWN]"
}

// ログを実際に出力する
func (self *Logger) log(level Level, v ...interface{}) {
	if self.level <= level {
		self.logger.Println(levelToString(level), fmt.Sprint(v...))
		if nil != self.writer {
			self.writer.Flush()
		}
	}
}

// ログを実際に出力する
func (self *Logger) logf(level Level, format string, args ...interface{}) {
	if self.level <= level {
		self.logger.Println(levelToString(level), fmt.Sprintf(format, args...))
		if nil != self.writer {
			self.writer.Flush()
		}
	}
}
