package main

import (
	"flag"
	"fmt"
	"github.com/uwork/gorond/config"
	"github.com/uwork/gorond/goron"
	"github.com/uwork/gorond/util"
	"github.com/uwork/gorond/webapi"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var version = "1.0.1"

func main() {
	configPath := flag.String("c", "/etc/goron.conf", "root config file")
	includeDir := flag.String("d", "/etc/goron.d/", "config directory")
	pidPath := flag.String("p", "/var/pid/gorond", "pid file")
	test := flag.Bool("t", false, "test config")
	version := flag.Bool("v", false, "show version")
	flag.Parse()

	var result int
	if *version {
		result = doVersion()
	} else if *test {
		result = doConfigTest(*configPath, *includeDir)
	} else {
		result = doMain(*configPath, *includeDir, *pidPath)
	}
	os.Exit(result)
}

func doConfigTest(configPath string, includeDir string) int {

	// load config test.
	_, err := config.LoadConfig(configPath, includeDir)
	if err != nil {
		log.Println(err)
		return -1
	}

	return 0
}

func doMain(configPath string, includeDir string, pidPath string) int {

	// load config.
	config, err := config.LoadConfig(configPath, includeDir)
	if err != nil {
		log.Println(err)
		return -1
	}

	// start goron.
	grn, err := goron.NewGorond(config)
	if err != nil {
		log.Println(err)
		return -1
	}

	// Goronデーモンの開始
	grn.Start()

	// API サーバの開始
	if grn.Config.Config.WebApi != "" {
		webapi.SetLogger(config.Config.ApiLog)
		server, err := webapi.NewWebApiServer(grn.Config.Config.WebApi, grn)
		if err != nil {
			log.Println(err)
			return -2
		}

		wc := make(chan os.Signal)
		wsc := make(chan error)
		signal.Notify(wc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		err = server.Start(wc, wsc)
		if err != nil {
			log.Println(err)
			return -2
		}
	}

	// create pid file.
	err = util.SavePidFile(pidPath)
	if err != nil {
		log.Println(err)
		return -3
	}
	defer os.Remove(pidPath)

	log.Println("wait for signal")

	// wait for terminate.
	c := make(chan os.Signal)
	sc := make(chan int, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go goron.WaitSignal(c, sc)

	return <-sc
}

// バージョンを表示
func doVersion() int {
	fmt.Println("gorond version", version)
	return 0
}
