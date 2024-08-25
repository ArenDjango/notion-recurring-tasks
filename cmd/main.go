package main

import (
	"github.com/ArenDjango/notion-recurring-tasks/internal/server"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
	})
	log.SetOutput(os.Stdout)
	level, err := log.ParseLevel(os.Getenv("LOGLVL"))
	if err != nil {
		level = log.InfoLevel
	}
	log.SetLevel(level)
}

func main() {
	log.Infof("Starting...")
	go print()

	srv := server.NewServer()

	srv.AddChecker(server.NewDefaultChecker("simple", func() error {
		return nil
	}))

	srv.Run()
}

func print() {
	for {
		time.Sleep(time.Second)
		log.Infof("Printing...")
	}
}
