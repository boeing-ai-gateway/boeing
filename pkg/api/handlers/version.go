package handlers

import (
	"context"
	"os"
	"slices"
	"strings"

	"github.com/boeing-ai-gateway/boeing/pkg/api"
	"github.com/boeing-ai-gateway/boeing/pkg/gateway/client"
	"github.com/boeing-ai-gateway/boeing/pkg/license"
	"github.com/boeing-ai-gateway/boeing/pkg/mcp"
	"github.com/boeing-ai-gateway/boeing/pkg/storage"
	"github.com/boeing-ai-gateway/boeing/pkg/version"
)

type SessionStore string

const (
	SessionStoreDB     SessionStore = "db"
	SessionStoreCookie SessionStore = "cookie"
)

func sessionStoreFromPostgresDSN(postgresDSN string) SessionStore {
	if postgresDSN != "" {
		return SessionStoreDB
	}
	return SessionStoreCookie
}

type VersionHandlerOptions struct {
	GatewayClient           *client.Client
	StorageClient           storage.Client
	LicenseProvider         *license.Provider
	PostgresDSN             string
	Engine                  string
	MCPNetworkPolicyEnabled bool
	MCPDefaultDenyAllEgress bool
	AuthEnabled             bool
	DisableUpdateCheck      bool
	MessagePoliciesEnabled  bool
	AgentsEnabled           bool
}

type VersionHandler struct {
	VersionHandlerOptions

	sessionStore SessionStore
}

func NewVersionHandler(_ context.Context, opts VersionHandlerOptions) (*VersionHandler, error) {
	return &VersionHandler{
		VersionHandlerOptions: opts,
		sessionStore:          sessionStoreFromPostgresDSN(opts.PostgresDSN),
	}, nil
}

func (v *VersionHandler) GetVersion(req api.Context) error {
	response, err := v.getVersionResponse(req.Context())
	if err != nil {
		return err
	}
	return req.Write(response)
}

func (v *VersionHandler) getVersionResponse(ctx context.Context) (map[string]any, error) {
	engine := v.Engine
	if mcp.IsKubernetesBackend(engine) {
		engine = mcp.RuntimeBackendKubernetes
	}

	violations, err := v.LicenseProvider.ConfiguredProviderViolations(ctx, v.StorageClient)
	if err != nil {
		return nil, err
	}

	values := map[string]any{
		"upgradeAvailable":             false,
		"latestVersion":                "",
		"boeing":                         version.Get().String(),
		"authEnabled":                  v.AuthEnabled,
		"sessionStore":                 v.sessionStore,
		"enterprise":                   v.LicenseProvider.HasValidLicense(),
		"licenseEntitlements":          v.LicenseProvider.Entitlements(),
		"engine":                       engine,
		"mcpNetworkPolicyEnabled":      v.MCPNetworkPolicyEnabled,
		"mcpDefaultDenyAllEgress":      v.MCPDefaultDenyAllEgress,
		"messagePoliciesEnabled":       v.MessagePoliciesEnabled,
		"agentsEnabled":                v.AgentsEnabled,
		"licenseEntitlementViolations": violations,
		"missingLicenseEntitlements":   missingEntitlements(violations),
	}

	if versions := os.Getenv("BOEING_SERVER_VERSIONS"); versions != "" {
		for pair := range strings.SplitSeq(versions, ",") {
			key, value, ok := strings.Cut(pair, "=")
			if !ok {
				continue
			}
			values[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}

	return values, nil
}

func missingEntitlements(violations []license.ProviderViolation) []string {
	seen := make(map[string]struct{})
	for _, violation := range violations {
		for _, entitlement := range violation.MissingEntitlements {
			seen[entitlement] = struct{}{}
		}
	}
	missing := make([]string, 0, len(seen))
	for entitlement := range seen {
		missing = append(missing, entitlement)
	}
	slices.Sort(missing)
	return missing
}
