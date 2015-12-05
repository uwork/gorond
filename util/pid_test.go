package util

import (
	"io/ioutil"
	"os"
	"strconv"
	"testing"
)

func TestSavePidFile(t *testing.T) {
	pidfile := "_pid"
	err := SavePidFile(pidfile)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(pidfile)

	data, err := ioutil.ReadFile(pidfile)
	if err != nil {
		t.Error(err)
	}

	intdata, err := strconv.Atoi(string(data))
	if err != nil {
		t.Error(err)
	}

	pid := os.Getpid()
	if pid != intdata {
		t.Errorf("pid (expected) %d != %d", pid, intdata)
	}
}

func TestExistsPidFile(t *testing.T) {
	pidfile := "_pid"

	// 現在のPIDと違うPIDのファイルを用意する
	err := ioutil.WriteFile(pidfile, []byte("1000000"), 0644)
	if err != nil {
		t.Error(err)
	}

	exists, err := ExistsPidFile(pidfile)
	if err != nil {
		t.Error(err)
	}
	if exists {
		t.Errorf("already exists current pid file.", pidfile)
	}

	// 現在のPIDファイルを作成する
	err = SavePidFile(pidfile)
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(pidfile)

	exists, err = ExistsPidFile(pidfile)
	if err != nil {
		t.Error(err)
	}
	if !exists {
		t.Errorf("current pid(%s) file not found.", pidfile)
	}
}
