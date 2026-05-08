package eventserver

import (
	"errors"
	"net/http"
	"sync"

	rrErrors "github.com/roadrunner-server/errors"
	"go.uber.org/zap"
	"gopkg.in/antage/eventsource.v1"
)

const (
	PluginName = "eventserver"
)

type Plugin struct {
	mu      sync.RWMutex
	cfg     *Config
	log     *zap.Logger
	metrics *statsExporter
	srv     *http.Server
	es      eventsource.EventSource
	eventId int
	errCh   chan error
	state   pluginState
	lastErr error
}

type pluginState uint8

const (
	stateStopped pluginState = iota
	stateRunning
	stateFailed
)

var errEventServerNotRunning = errors.New("event server is not running")

func (s *Plugin) Init(cfg Configurer, log Logger) error {
	const op = rrErrors.Op("event_server_plugin_init")
	if !cfg.Has(PluginName) {
		return rrErrors.E(op, rrErrors.Disabled)
	}

	err := cfg.UnmarshalKey(PluginName, &s.cfg)
	if err != nil {
		return rrErrors.E(op, err)
	}

	s.cfg.InitDefaults()

	s.log = log.NamedLogger(PluginName)

	s.eventId = 0
	s.metrics = newStatsExporter()
	s.state = stateStopped
	s.lastErr = nil

	return nil
}

func (s *Plugin) Name() string {
	return PluginName
}

func (s *Plugin) RPC() any {
	return &rpc{plugin: s, log: s.log}
}

func newEventSource() eventsource.EventSource {
	return eventsource.New(
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
