package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
)

// ServerControl is used to access informations and control a running http server
type ServerControl struct {
	// HTTP Server controlled by this instance
	Server *http.Server
	// Channel setup to listen for SIGINT
	Interrupted chan os.Signal
	// Error channel used by the running server if any error is detected during launch or while running
	Errors chan error
	status *serverStatus
}

// Status function load current server status
func (c *ServerControl) Status() Status {
	return c.status.load()
}

// Wait will wait for the first server error, or a SIGINT signal. If a SIGINT is received, the control server will be shutdown
func (c *ServerControl) Wait() error {
	select {
	case err := <-c.Errors:
		if nil != err {
			return err
		}
	case <-c.Interrupted:
		if err := c.Shutdown(5 * time.Second); nil != err {
			return err
		}
	}
	return nil
}

// Shutdown will shutdown the controlled server
func (c ServerControl) Shutdown(timeout time.Duration) error {
	c.status.set(Closing)
	defer c.status.set(Closed)

	timeoutCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := c.Server.Shutdown(timeoutCtx); nil != err {
		c.status.set(Error)
		return fmt.Errorf("shutdown http server: %w", err)
	}
	return nil
}
