package eventServer

import (
	"context"
	"github.com/roadrunner-server/errors"
	"go.uber.org/zap"
	"gopkg.in/antage/eventsource.v1"
	"net/http"
)

func (s *Plugin) Serve() chan error {
	const op = errors.Op("event_server_serve")
	errCh := make(chan error, 1)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.log.Info("Starting event server...", zap.String("address", s.cfg.Address))

	mux := http.NewServeMux()
	mux.Handle("/events", s.es)
	s.srv = &http.Server{Addr: s.cfg.Address, Handler: mux}
	err := s.srv.ListenAndServe()
	if err != nil {
		errCh <- errors.E(op, err)
		return errCh
	}

	return errCh
}

func (s *Plugin) Stop(ctx context.Context) error {
	s.log.Info("Stopping event server...")
	s.eventId = 0
	s.es.Close()
	err := s.srv.Shutdown(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *Plugin) Reset() error {
	const op = errors.Op("event_server_reset")
	s.mu.Lock()
	defer s.mu.Unlock()

	s.log.Info("Resetting event server...")

	// Reset the event ID and close the existing event source
	s.eventId = 0
	if s.es != nil {
		s.es.Close()
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
	} else {
		return errors.E(op, errors.Disabled, "event source is not initialized")
	}

	// Stop the existing server if it is running
	if s.srv != nil {
		err := s.srv.Shutdown(context.Background())
		if err != nil {
			return errors.E(op, err)
		}
	}

	// Start a new server instance
	mux := http.NewServeMux()
	mux.Handle("/events", s.es)
	s.srv = &http.Server{Addr: s.cfg.Address, Handler: mux}
	err := s.srv.ListenAndServe()
	if err != nil {
		return errors.E(op, err)
	}

	return nil
}
