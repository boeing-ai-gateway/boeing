package services

import (
	"context"
)

// Otel is a no-op OpenTelemetry placeholder.
// No traces, metrics, or logs are exported to any external service.
type Otel struct{}

func (s *Otel) Shutdown(_ context.Context) error {
	return nil
}

// newOtel returns a no-op Otel instance. No external telemetry is configured or exported.
func newOtel(_ context.Context) (*Otel, error) {
	return &Otel{}, nil
}
