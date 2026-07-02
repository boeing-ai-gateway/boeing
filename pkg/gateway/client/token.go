package client

import (
	"context"

	"github.com/boeing-ai-gateway/boeing/pkg/gateway/types"
)

// CreateTokenRequest creates a new token request in the database.
func (c *Client) CreateTokenRequest(ctx context.Context, tr *types.TokenRequest) error {
	return c.db.WithContext(ctx).Create(tr).Error
}
