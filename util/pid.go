package util

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

func SavePidFile(pidfile string) error {
	exists, err := ExistsPidFile(pidfile)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("pidfile already exists.")
	}

	piddir := filepath.Dir(pidfile)
	if !ExistsFile(piddir) {
		err := os.Mkdir(piddir, 0755)
		if err != nil {
			return err
		}
	}

	pid := strconv.Itoa(syscall.Getpid())
	err = ioutil.WriteFile(pidfile, []byte(pid), 0644)
	if err != nil {
		return err
	}

	return nil
}

// 既存のPIDファイルがこのプロセスのPIDと同じPIDか確認する
func ExistsPidFile(pidfile string) (bool, error) {
	if !ExistsFile(pidfile) {
		return false, nil
	}

	fpid, err := ioutil.ReadFile(pidfile)
	if err != nil {
		return false, err
	}

	ipid, err := strconv.Atoi(string(fpid))
	if err != nil {
		return false, err
	}

	pid := syscall.Getpid()
	if pid != ipid {
		return false, nil
	} else {
		return true, nil
	}
}
