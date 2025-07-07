package eventServer

import (
	"context"
	"github.com/roadrunner-server/errors"
	"gopkg.in/antage/eventsource.v1"
	"net/http"
)

func (s *Plugin) Serve() chan error {
	const op = errors.Op("custom_plugin_serve")
	errCh := make(chan error, 1)

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
	http.Handle("/events", s.es)

	err := http.ListenAndServe(s.cfg.Address, nil)
	if err != nil {
		errCh <- errors.E(op, err)
		return errCh
	}

	return nil
}

func (s *Plugin) Stop(ctx context.Context) error {
	s.es.Close()
	return nil
}
