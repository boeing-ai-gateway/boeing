package mcp

import (
	"context"

	nmcp "github.com/boeing-ai-gateway/boeingbot/pkg/mcp"
)

func (sm *SessionManager) ServerCapabilities(ctx context.Context, serverConfig ServerConfig) (nmcp.ServerCapabilities, error) {
	client, err := sm.clientForServer(ctx, serverConfig)
	if err != nil {
		return nmcp.ServerCapabilities{}, err
	}

	return client.Session.InitializeResult.Capabilities, nil
}
