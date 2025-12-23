package eventserver

import (
	"github.com/roadrunner-server/api/v4/plugins/v1/status"
	"go.uber.org/zap"
)

type Logger interface {
	NamedLogger(name string) *zap.Logger
}

type Configurer interface {
	// UnmarshalKey takes a single key and unmarshal it into a Struct.
	UnmarshalKey(name string, out any) error
	// Has checks if config section exists.
	Has(name string) bool
}

// Checker interface used to get the latest status from the plugin
type Checker interface {
	Status() (*status.Status, error)
	Name() string
}

type Readiness interface {
	Ready() (*status.Status, error)
	Name() string
}
