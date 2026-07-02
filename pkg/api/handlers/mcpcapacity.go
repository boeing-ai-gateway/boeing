package handlers

import (
	"errors"

	"github.com/boeing-ai-gateway/boeing/apiclient/types"
	"github.com/boeing-ai-gateway/boeing/pkg/api"
	"github.com/boeing-ai-gateway/boeing/pkg/mcp"
)

type MCPCapacityHandler struct {
	mcpSessionManager *mcp.SessionManager
}

func NewMCPCapacityHandler(mcpSessionManager *mcp.SessionManager) *MCPCapacityHandler {
	return &MCPCapacityHandler{
		mcpSessionManager: mcpSessionManager,
	}
}

// GetCapacity returns capacity information for the MCP namespace.
// This endpoint is admin/owner-only.
func (h *MCPCapacityHandler) GetCapacity(req api.Context) error {
	info, err := h.mcpSessionManager.GetCapacityInfo(req.Context())
	if err != nil {
		// If backend doesn't support capacity info (e.g., Docker), return empty info
		var notSupported *mcp.ErrNotSupportedByBackend
		if errors.As(err, &notSupported) {
			return req.Write(types.MCPCapacityInfo{
				Error: notSupported.Error(),
			})
		}
		return err
	}

	return req.Write(info)
}
