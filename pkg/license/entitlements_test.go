package license

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMissingEntitlementsAlwaysEmpty(t *testing.T) {
	provider, err := NewProvider(context.Background(), nil, Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	missing := provider.MissingEntitlements([]string{"ENTITLED", "MISSING", "ANYTHING"})
	if len(missing) != 0 {
		t.Fatalf("MissingEntitlements() = %v, want empty (all granted)", missing)
	}
}

func TestRequireEntitlementsAlwaysNil(t *testing.T) {
	provider, err := NewProvider(context.Background(), nil, Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := provider.RequireEntitlements([]string{"ANY_ENTITLEMENT"}); err != nil {
		t.Fatalf("RequireEntitlements() error = %v, want nil", err)
	}
}

func TestProviderEntitlementGateAlwaysPasses(t *testing.T) {
	provider, err := NewProvider(context.Background(), nil, Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gate := NewProviderEntitlementGate(provider, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/llm-proxy/test", nil)
	if err := gate.Check(req); err != nil {
		t.Fatalf("Check() error = %v, want nil (no gating)", err)
	}
}

func TestConfiguredProviderViolationsAlwaysEmpty(t *testing.T) {
	provider, err := NewProvider(context.Background(), nil, Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	violations, err := provider.ConfiguredProviderViolations(context.Background(), nil)
	if err != nil {
		t.Fatalf("ConfiguredProviderViolations() error = %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("ConfiguredProviderViolations() = %v, want empty", violations)
	}
}
