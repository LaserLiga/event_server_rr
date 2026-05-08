package eventserver

import (
	rrErrors "github.com/roadrunner-server/errors"
	"go.uber.org/zap"
	"strconv"
)

type rpc struct {
	plugin *Plugin
	log    *zap.Logger
}

type EventMessage struct {
	Type    string `json:"type"`
	Message string `json:"data"`
}

func (r *rpc) TriggerEvent(input EventMessage, output *int) error {
	const op = rrErrors.Op("event_server_trigger_event")

	r.plugin.mu.Lock()
	defer r.plugin.mu.Unlock()

	if r.plugin.state != stateRunning || r.plugin.es == nil {
		return rrErrors.E(op, errEventServerNotRunning)
	}

	r.plugin.es.SendEventMessage(input.Type, "event", strconv.Itoa(r.plugin.eventId))
	r.plugin.eventId++

	if input.Message != "" {
		r.plugin.es.SendEventMessage(input.Message, "message", strconv.Itoa(r.plugin.eventId))
		r.plugin.eventId++
	}

	*output = r.plugin.eventId
	r.log.Info("Event triggered", zap.String("type", input.Type), zap.String("message", input.Message), zap.Int("eventId", *output))
	if r.plugin.metrics != nil {
		r.plugin.metrics.CountEvents()
	}
	return nil
}
