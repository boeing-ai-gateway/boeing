package license

import (
	"context"
	"net/http"

	"github.com/boeing-ai-gateway/boeing/apiclient/types"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var entitlementPathsToGate = []string{
	"/mcp-connect/{mcp_id}",
	"/mcp-connect/{mcp_id}/",
	"GET /oauth/authorize",
	"GET /oauth/authorize/",
	"GET /oauth/consent/",
	"POST /oauth/consent/",
	"GET /oauth/complete/",
	"GET /oauth/mcp/callback/",
	"POST /oauth/",
	"PUT /oauth/",
	"GET /api/oauth/composite/{mcp_id}",
	"/api/llm-proxy/",
	"/api/skills",
	"/api/skills/",
	"POST /api/devices/scans",
}

// ProviderViolation describes a configured provider that requires license entitlements
// that are not currently available.
type ProviderViolation struct {
	Type                 string   `json:"type"`
	Namespace            string   `json:"namespace"`
	Name                 string   `json:"name"`
	RequiredEntitlements []string `json:"requiredEntitlements"`
	MissingEntitlements  []string `json:"missingEntitlements"`
}

type ProviderMeta struct {
	RequiredEntitlements []string                               `json:"requiredEntitlements"`
	EnvVars              []types.ProviderConfigurationParameter `json:"envVars"`
}

type ProviderEntitlementGate struct {
	licenseProvider *Provider
	client          kclient.Client
	mux             *http.ServeMux
}

func NewProviderEntitlementGate(licenseProvider *Provider, client kclient.Client) *ProviderEntitlementGate {
	mux := http.NewServeMux()
	for _, path := range entitlementPathsToGate {
		mux.Handle(path, (*fake)(nil))
	}

	return &ProviderEntitlementGate{
		licenseProvider: licenseProvider,
		client:          client,
		mux:             mux,
	}
}

// Check always returns nil — no entitlement gating is enforced.
func (g *ProviderEntitlementGate) Check(_ *http.Request) error {
	return nil
}

func (g *ProviderEntitlementGate) requiresProviderEntitlements(req *http.Request) bool {
	_, pattern := g.mux.Handler(req)
	return pattern != ""
}

// MissingEntitlements always returns nil — all entitlements are granted.
func (p *Provider) MissingEntitlements(_ []string) []string {
	return nil
}

// RequireEntitlements always returns nil — all entitlements are granted.
func (p *Provider) RequireEntitlements(_ []string) error {
	return nil
}

// ConfiguredProviderViolations always returns no violations.
func (p *Provider) ConfiguredProviderViolations(_ context.Context, _ kclient.Client) ([]ProviderViolation, error) {
	return nil, nil
}

// fake is a fake handler that does fake things
type fake struct{}

func (f *fake) ServeHTTP(http.ResponseWriter, *http.Request) {}
