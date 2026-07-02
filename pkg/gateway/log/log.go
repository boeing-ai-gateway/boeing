//nolint:revive
package log

import "github.com/boeing-ai-gateway/boeing/logger"

func NewWithID(id string) *logger.Logger {
	log := logger.New("gateway")
	if id != "" {
		return log.Fields("req_id", id)
	}
	return &log
}
