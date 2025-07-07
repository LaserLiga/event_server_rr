package eventServer

import (
	"context"
	"github.com/roadrunner-server/errors"
	"go.uber.org/zap"
	"net/http"
)

func (s *Plugin) Serve() chan error {
	const op = errors.Op("custom_plugin_serve")
	errCh := make(chan error, 1)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.log.Info("Starting event server...", zap.String("address", s.cfg.Address))
	http.Handle("/events", s.es)

	err := http.ListenAndServe(s.cfg.Address, nil)
	if err != nil {
		errCh <- errors.E(op, err)
		return errCh
	}

	return errCh
}

func (s *Plugin) Stop(ctx context.Context) error {
	s.log.Info("Stopping event server...")
	s.es.Close()
	return nil
}
