package main

import (
	"flag"
	"fmt"
	"gorond/config"
	"gorond/goron"
	"gorond/webapi"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var version = "1.0"

func main() {
	configPath := flag.String("c", "/etc/goron.conf", "root config file")
	includeDir := flag.String("d", "/etc/goron.d/", "config directory")
	version := flag.Bool("v", false, "show version")
	flag.Parse()

	var result int
	if *version {
		result = doVersion()
	} else {
		result = doMain(*configPath, *includeDir)
	}
	os.Exit(result)
}

func doMain(configPath string, includeDir string) int {

	// load config.
	config, err := config.LoadConfig(configPath, includeDir)
	if err != nil {
		log.Fatal(err)
		return -1
	}

	// start goron.
	grn, err := goron.NewGorond(config)
	if err != nil {
		log.Fatal(err)
		return -1
	}

	// Goronデーモンの開始
	_, err = goron.StartAutoReload(grn, configPath, includeDir)
	if err != nil {
		return -1
	}
	grn.Start()

	// API サーバの開始
	webapi.SetLogger(config.Config.ApiLog)
	server, err := webapi.NewWebApiServer(grn.Config.Config.WebApi, grn)
	if err != nil {
		log.Fatal(err)
		return -2
	}

	wc := make(chan os.Signal)
	wsc := make(chan error)
	signal.Notify(wc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	err = server.Start(wc, wsc)
	if err != nil {
		log.Fatal(err)
		return -2
	}

	log.Println("wait for signal")

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
