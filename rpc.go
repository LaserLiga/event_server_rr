package eventserver

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

type eventMessage struct {
	Type    string `json:"type"`
	Message string `json:"data"`
}

func (r *rpc) TriggerEvent(input eventMessage, output *int) error {
	r.mx.Lock()

	r.plugin.es.SendEventMessage(input.Type, "event", strconv.Itoa(r.plugin.eventId))
	r.plugin.eventId++

	if input.Message != "" {
		r.plugin.es.SendEventMessage(input.Message, "message", strconv.Itoa(r.plugin.eventId))
		r.plugin.eventId++
	}
	r.mx.Unlock()

	*output = r.plugin.eventId
	r.log.Info("Event triggered", zap.String("type", input.Type), zap.String("message", input.Message), zap.Int("eventId", r.plugin.eventId))
	if r.plugin.metrics != nil {
		r.plugin.metrics.CountEvents()
	}
	return nil
}
