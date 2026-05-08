package eventserver

import (
	"context"
	er "errors"
	"net"
	"net/http"
	"time"

	rrErrors "github.com/roadrunner-server/errors"
	"go.uber.org/zap"
)

const resetShutdownTimeout = 5 * time.Second

func (s *Plugin) Serve() chan error {
	const op = rrErrors.Op("event_server_serve")

	s.mu.Lock()
	defer s.mu.Unlock()

	s.log.Info("Starting event server...", zap.String("address", s.cfg.Address))

	if s.state == stateRunning {
		return s.errCh
	}

	s.errCh = make(chan error, 1)
	if err := s.startServerLocked(op); err != nil {
		s.errCh <- err
	}

	return s.errCh
}

func (s *Plugin) Stop(ctx context.Context) error {
	const op = rrErrors.Op("event_server_stop")

	s.mu.Lock()
	srv := s.srv
	es := s.es
	s.srv = nil
	s.es = nil
	s.eventId = 0
	s.state = stateStopped
	s.lastErr = nil
	s.mu.Unlock()

	s.log.Info("Stopping event server...")

	if es != nil {
		es.Close()
	}

	if srv != nil {
		if err := srv.Shutdown(ctx); err != nil {
			return rrErrors.E(op, err)
		}
	}

	return nil
}

func (s *Plugin) Reset() error {
	const op = rrErrors.Op("event_server_reset")

	s.mu.Lock()
	srv := s.srv
	es := s.es
	s.srv = nil
	s.es = nil
	s.eventId = 0
	s.state = stateStopped
	s.lastErr = nil
	s.mu.Unlock()

	s.log.Info("Resetting event server...")

	if es != nil {
		es.Close()
	}

	if srv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), resetShutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return rrErrors.E(op, err)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.errCh == nil {
		s.errCh = make(chan error, 1)
	}
	if err := s.startServerLocked(op); err != nil {
		select {
		case s.errCh <- err:
		default:
		}
		return err
	}

	s.log.Info("Event server successfully reset")
	return nil
}

func (s *Plugin) startServerLocked(op rrErrors.Op) error {
	if s.state == stateRunning {
		return rrErrors.E(op, rrErrors.Str("event server is already running"))
	}

	ln, err := net.Listen("tcp", s.cfg.Address)
	if err != nil {
		wrapped := rrErrors.E(op, err)
		s.state = stateFailed
		s.lastErr = wrapped
		return wrapped
	}

	s.es = newEventSource()
	mux := http.NewServeMux()
	mux.Handle("/events", s.es)
	s.srv = &http.Server{Addr: s.cfg.Address, Handler: mux}
	s.state = stateRunning
	s.lastErr = nil

	srv := s.srv
	errCh := s.errCh
	go func() {
		if err := srv.Serve(ln); err != nil && !er.Is(err, http.ErrServerClosed) {
			wrapped := rrErrors.E(op, err)
			s.mu.Lock()
			if s.srv == srv {
				s.state = stateFailed
				s.lastErr = wrapped
			}
			s.mu.Unlock()

			select {
			case errCh <- wrapped:
			default:
			}
		}
	}()

	return nil
}
