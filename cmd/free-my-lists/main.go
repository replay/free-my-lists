package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/replay/free-my-lists/pkg/config"
	"github.com/replay/free-my-lists/pkg/web"
)

func main() {
	var configFile string

	flag.StringVar(&configFile, "config.file", "free-my-lists.json", "Path to the config file")
	flag.Parse()

	if len(configFile) == 0 {
		log.Fatal("no config file specified")
	}

	conf, err := config.GetConfig(configFile)
	if err != nil {
		panic(err)
	}

	w := web.New(conf)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		w.Shutdown()
		os.Exit(1)
	}()
}
