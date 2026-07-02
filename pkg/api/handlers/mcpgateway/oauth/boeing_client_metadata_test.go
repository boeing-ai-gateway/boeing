package oauth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/boeing-ai-gateway/boeing/apiclient/types"
	"github.com/boeing-ai-gateway/boeing/pkg/api"
	"github.com/boeing-ai-gateway/boeing/pkg/system"
)

func TestBoeingClientIDMetadata(t *testing.T) {
	t.Parallel()

	const baseURL = "https://boeing.example"
	h := &handler{baseURL: baseURL}
	req := httptest.NewRequest(http.MethodGet, system.OAuthClientIDMetadataPath, nil)
	rec := httptest.NewRecorder()

	if err := h.boeingClientIDMetadata(api.Context{
		ResponseWriter: rec,
		Request:        req,
	}); err != nil {
		t.Fatalf("serve metadata: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var got struct {
		types.OAuthClientManifest
		ClientID string `json:"client_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}

	if got.ClientID != system.OAuthClientIDMetadataURL(baseURL) {
		t.Fatalf("expected client_id %q, got %q", system.OAuthClientIDMetadataURL(baseURL), got.ClientID)
	}
	if got.ClientName != "Boeing MCP Gateway" {
		t.Fatalf("unexpected client_name: %s", got.ClientName)
	}
	if got.TokenEndpointAuthMethod != "none" {
		t.Fatalf("expected token_endpoint_auth_method none, got %q", got.TokenEndpointAuthMethod)
	}
	if len(got.RedirectURIs) != 1 || got.RedirectURIs[0] != system.MCPOAuthCallbackURL(baseURL) {
		t.Fatalf("unexpected redirect_uris: %#v", got.RedirectURIs)
	}
}
