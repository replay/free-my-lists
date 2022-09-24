package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/replay/free-my-lists/pkg/config"
	"github.com/replay/free-my-lists/pkg/web"
)

func main() {

	if len(os.Args) < 2 {
		panic("missing config file")
	}

	conf, err := config.GetConfig(os.Args[1])
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
