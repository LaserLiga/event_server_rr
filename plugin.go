package eventServer

import (
	"gopkg.in/antage/eventsource.v1"
	"net/http"
	"sync"
)
import "go.uber.org/zap"
import "github.com/roadrunner-server/errors"
import _ "gopkg.in/antage/eventsource.v1"

const (
	PluginName = "event_server"

	// v2.7 and newer config key
	cfgKey string = "config"
)

type Plugin struct {
	mu      sync.RWMutex
	cfg     *Config
	log     *zap.Logger
	metrics *statsExporter
	srv     *http.Server
	es      eventsource.EventSource
	eventId int
}

func (s *Plugin) Init(cfg Configurer, log Logger) error {
	const op = errors.Op("event_server_plugin_init")
	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(PluginName, &s.cfg)
	if err != nil {
		return errors.E(op, err)
	}

	s.cfg.InitDefaults()

	s.log = new(zap.Logger)
	s.log = log.NamedLogger(PluginName)

	s.eventId = 0
	s.metrics = newStatsExporter()
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

	return nil
}
