package server

import (
	"context"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type engineCfg struct {
	DebugPort       string        `json:"debug_port" envconfig:"DEBUG_PORT" default:"804"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout" envconfig:"SHUTDOWN_TIMEOUT" default:"30s"`
}

type Server struct {
	*DebugServer

	shutdownTimeout time.Duration
}

func (s *Server) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	log.Infof("Run server")
	go func() {
		err := s.DebugServer.Run(ctx)
		if err != nil && err != http.ErrServerClosed {
			log.Errorf("Error run debug server: %s", err)
		}
	}()
	s.SetReady(true)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit
	log.Infof("Shutting down server")
	s.Shutdown(ctx)
	cancel()
	log.Infof("Shutdown server")
}

func NewServer() Server {
	var cfg engineCfg

	err := envconfig.Process("", &cfg)
	if err != nil {
		log.Errorf("Error processing environment variables: %s", err)
		panic(err)
	}

	debug := NewDebugServer(cfg.DebugPort)

	s := Server{DebugServer: debug, shutdownTimeout: cfg.ShutdownTimeout}

	return s
}
