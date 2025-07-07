package eventServer

import (
	"go.uber.org/zap"
	"strconv"
	"sync"
)

type rpc struct {
	plugin *Plugin
	log    *zap.Logger
	mx     sync.Mutex
}

func (p *Plugin) RPC() any {
	return &rpc{plugin: p, log: p.log, mx: sync.Mutex{}}
}

type eventMessage struct {
	Type    string `json:"type"`
	Message string `json:"data"`
}

func (s *rpc) TriggerEvent(input eventMessage, output *int) error {
	s.mx.Lock()

	s.plugin.es.SendEventMessage(input.Type, "event", strconv.Itoa(s.plugin.eventId))
	s.plugin.eventId++

	if input.Message != "" {
		s.plugin.es.SendEventMessage(input.Message, "message", strconv.Itoa(s.plugin.eventId))
		s.plugin.eventId++
	}
	s.mx.Unlock()

	*output = s.plugin.eventId
	s.log.Info("Event triggered", zap.String("type", input.Type), zap.String("message", input.Message), zap.Int("eventId", s.plugin.eventId))
	if s.plugin.metrics != nil {
		s.plugin.metrics.CountEvents()
	}
	return nil
}
