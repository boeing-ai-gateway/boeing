package oauth

import (
	"github.com/boeing-ai-gateway/boeing/apiclient/types"
	"github.com/boeing-ai-gateway/boeing/pkg/api"
	"github.com/boeing-ai-gateway/boeing/pkg/system"
)

func (h *handler) boeingClientIDMetadata(req api.Context) error {
	return req.Write(clientIDMetadataDocument{
		ClientID: system.OAuthClientIDMetadataURL(h.baseURL),
		OAuthClientManifest: types.OAuthClientManifest{
			RedirectURIs:            []string{system.MCPOAuthCallbackURL(h.baseURL)},
			TokenEndpointAuthMethod: "none",
			GrantTypes:              []string{"authorization_code", "refresh_token"},
			ResponseTypes:           []string{"code"},
			ClientName:              "Boeing MCP Gateway",
			ClientURI:               h.baseURL,
			SoftwareID:              "boeing",
		},
	})
}
