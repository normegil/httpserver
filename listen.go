package httpserver

import (
	"fmt"
	"github.com/normegil/closer"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
)

// Status represent the current HTTP server status
type Status int

const (
	NotStarted Status = iota
	Starting
	Listening
	Closing
	Closed
	Error
)

// Server is an object used to launch an HTTP server
type Server struct {
	// Address controls the address used by the HTTP server
	Address string
	// Port controls the address used by the HTTP server
	Port int
	// Handler will be passed to the HTTP server to handle HTTP calls
	Handler http.Handler
	// Closer will be used to handler errors in ressource closing in defer statement. By default, the error is discarded
	Closer closer.CloseErrorHandler
}

func (s Server) init() {
	if nil == s.Closer {
		s.Closer = closer.DiscardErrorHandler{}
	}
}

// Listen launch the HTTP server in a separate goroutine and return an instance to control the server lifecycle.
func (s Server) Listen() *ServerControl {
	s.init()

	stopHTTPServer := make(chan os.Signal, 1)
	signal.Notify(stopHTTPServer, os.Interrupt)

	status := &serverStatus{
		lock:   sync.RWMutex{},
		status: NotStarted,
	}

	srv := &http.Server{Handler: s.Handler}

	errs := make(chan error, 1)
	go func() {
		status.set(Starting)
		httpAddr := s.Address + ":" + strconv.Itoa(s.Port)
		l, err := net.Listen("tcp", httpAddr)
		if err != nil {
			status.set(Error)
			errs <- fmt.Errorf("listening to %s: %w", httpAddr, err)
			return
		}

		status.set(Listening)

		if err := srv.Serve(l); nil != err {
			if http.ErrServerClosed != err {
				status.set(Error)
				errs <- fmt.Errorf("serving on %s: %w", httpAddr, err)
				close(errs)
			}
		}
	}()

	return &ServerControl{
		status:      status,
		Interrupted: stopHTTPServer,
		Errors:      errs,
		Server:      srv,
	}
}

type serverStatus struct {
	lock   sync.RWMutex
	status Status
}

func (s *serverStatus) load() Status {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.status
}

func (s *serverStatus) set(status Status) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if s.status != Error {
		s.status = status
	}
}
