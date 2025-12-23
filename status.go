package eventserver

import (
	"github.com/roadrunner-server/api/v4/plugins/v1/status"
	"net/http"
)

func (s *Plugin) Status() (*status.Status, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if the server is running
	if s.srv != nil && s.es != nil {
		return &status.Status{
			Code: http.StatusOK,
		}, nil
	}

	return &status.Status{
		Code: http.StatusServiceUnavailable,
	}, nil
}

func (s *Plugin) Ready() (*status.Status, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Check if the server is running
	if s.srv != nil && s.es != nil {
		return &status.Status{
			Code: http.StatusOK,
		}, nil
	}

	return &status.Status{
		Code: http.StatusServiceUnavailable,
	}, nil
}
