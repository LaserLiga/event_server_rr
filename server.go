package eventserver

import (
	"context"
	er "errors"
	"github.com/roadrunner-server/errors"
	"go.uber.org/zap"
	"gopkg.in/antage/eventsource.v1"
	"net/http"
)

func (s *Plugin) Serve() chan error {
	const op = errors.Op("event_server_serve")
	s.errCh = make(chan error, 1)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.log.Info("Starting event server...", zap.String("address", s.cfg.Address))

	// Ensure everything is properly initialized
	if s.es == nil {
		s.es = eventsource.New(
			eventsource.DefaultSettings(),
			func(req *http.Request) [][]byte {
				return [][]byte{
					[]byte("Connection: keep-alive"),
					[]byte("Cache-Control: no-cache"),
					[]byte("Access-Control-Allow-Origin: *"),
					[]byte("Access-Control-Allow-Methods: GET, POST, OPTIONS"),
				}
			},
		)
	}

	mux := http.NewServeMux()
	mux.Handle("/events", s.es)
	s.srv = &http.Server{Addr: s.cfg.Address, Handler: mux}

	go func() {
		err := s.srv.ListenAndServe()
		if err != nil && !er.Is(err, http.ErrServerClosed) {
			s.errCh <- errors.E(op, err)
		}
	}()

	return s.errCh
}

func (s *Plugin) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	const op = errors.Op("event_server_stop")

	s.log.Info("Stopping event server...")

	// Reset event ID
	s.eventId = 0

	// Close event source if it exists
	if s.es != nil {
		s.es.Close()
		s.es = nil
	}

	// Shutdown HTTP server if it exists
	if s.srv != nil {
		err := s.srv.Shutdown(ctx)
		if err != nil {
			return errors.E(op, err)
		}
		s.srv = nil
	}

	return nil
}

func (s *Plugin) Reset() error {
	const op = errors.Op("event_server_reset")

	s.mu.Lock()
	defer s.mu.Unlock()

	s.log.Info("Resetting event server...")

	// Reset event ID
	s.eventId = 0

	// Close existing event source
	if s.es != nil {
		s.es.Close()
	}

	// Shutdown HTTP server if it exists
	if s.srv != nil {
		err := s.srv.Shutdown(context.Background())
		if err != nil {
			return errors.E(op, err)
		}
	}

	// Create a new event source
	s.es = eventsource.New(
		eventsource.DefaultSettings(),
		func(req *http.Request) [][]byte {
			return [][]byte{
				[]byte("Connection: keep-alive"),
				[]byte("Cache-Control: no-cache"),
				[]byte("Access-Control-Allow-Origin: *"),
				[]byte("Access-Control-Allow-Methods: GET, POST, OPTIONS"),
			}
		},
	)

	// Start a new HTTP server
	mux := http.NewServeMux()
	mux.Handle("/events", s.es)
	s.srv = &http.Server{Addr: s.cfg.Address, Handler: mux}

	go func() {
		err := s.srv.ListenAndServe()
		if err != nil && !er.Is(err, http.ErrServerClosed) {
			s.errCh <- errors.E(op, err)
		}
	}()

	s.log.Info("Event server successfully reset")
	return nil
}
