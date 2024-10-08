package server

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
)

type DebugServer struct {
	PORT string `json:"port" envconfig:"PORT" default:"8084"`

	engine *echo.Echo
	m      sync.Mutex

	ready    bool
	checkers []Checker

	shutdownTimeout time.Duration
}

func NewDebugServer(port string) *DebugServer {
	e := echo.New()
	e.Debug = false
	e.HideBanner = true
	e.HidePort = true

	debug := &DebugServer{
		PORT:            port,
		engine:          e,
		m:               sync.Mutex{},
		ready:           false,
		checkers:        make([]Checker, 0),
		shutdownTimeout: time.Second * 30,
	}

	e.GET("/ready", debug.Ready)
	e.GET("/live", debug.Live)

	return debug
}

func (d *DebugServer) SetReady(ready bool) {
	d.m.Lock()
	defer d.m.Unlock()
	if ready {
		log.Infof("Server is ready")
	} else {
		log.Infof("Server is not ready")
	}
}

func (d *DebugServer) AddChecker(checker Checker) {
	log.Debugf("Adding checker: %v", checker.Name())
	d.m.Lock()
	defer d.m.Unlock()
	d.checkers = append(d.checkers, checker)
}

func (d *DebugServer) AddCheckers(checkers []Checker) {
	for _, checker := range checkers {
		d.AddChecker(checker)
	}
}

func (d *DebugServer) Ready(c echo.Context) error {
	d.m.Lock()
	defer d.m.Unlock()
	return c.JSON(http.StatusOK, map[string]interface{}{
		"ready": d.ready,
	})
}

func (d *DebugServer) Live(c echo.Context) error {
	d.m.Lock()
	defer d.m.Unlock()
	for i := range d.checkers {
		if err := d.checkers[i].Check(); err != nil {
			log.Errorf("Checker %s failed: %s", d.checkers[i].Name(), err)
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}
	log.Infof("Checkers passed")
	return c.String(http.StatusOK, "ok")
}

type Checker interface {
	Check() error
	Name() string
}

type DefaultChecker struct {
	CheckFunc func() error `json:"check"`
	NameCheck string       `json:"name"`
}

func (c *DefaultChecker) Check() error {
	return c.CheckFunc()
}

func (c *DefaultChecker) Name() string {
	return c.NameCheck
}

func NewDefaultChecker(name string, checkFunc func() error) *DefaultChecker {
	return &DefaultChecker{
		CheckFunc: checkFunc,
		NameCheck: name,
	}
}

func (d *DebugServer) Run(ctx context.Context) error {
	log.Infof("Run debug server at %s", time.Now())
	var err error
	func() {
		err = d.engine.Start(fmt.Sprintf(":%s", d.PORT))
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err != nil {
				d.SetReady(false)
				return err
			}
		}
	}
}

func (d *DebugServer) Shutdown(ctx context.Context) {
	log.Infof("Shutdown debug server at %s", time.Now())
	d.SetReady(false)
	ctxShutDown, cancel := context.WithTimeout(context.Background(), d.shutdownTimeout)
	defer cancel()

	errShutDown := d.engine.Shutdown(ctxShutDown)
	if errShutDown != nil {
		log.Panic("Shutdown debug server error: ", errShutDown)
	}

	log.Info("Debug server shutdown graceful")
}

func (d *DebugServer) setShutDownTimeout(timeout time.Duration) {
	d.shutdownTimeout = timeout
}
