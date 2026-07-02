package license

import (
	"context"
	"testing"
)

func TestNewProviderAlwaysValid(t *testing.T) {
	provider, err := NewProvider(context.Background(), nil, Config{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if provider == nil {
		t.Fatal("expected provider to be created")
	}
	if !provider.HasValidLicense() {
		t.Fatal("expected license to always be valid")
	}
}

func TestNewProviderWithKey(t *testing.T) {
	provider, err := NewProvider(context.Background(), nil, Config{
		LicenseKey: "test-key",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !provider.HasValidLicense() {
		t.Fatal("expected license to be valid")
	}
	if provider.LicenseKey() != "test-key" {
		t.Fatalf("expected license key to be 'test-key', got %q", provider.LicenseKey())
	}
	if !provider.LicenseKeyViaConfiguration() {
		t.Fatal("expected LicenseKeyViaConfiguration to be true")
	}
}

func TestEntitlementsAlwaysGranted(t *testing.T) {
	provider, err := NewProvider(context.Background(), nil, Config{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !provider.hasEntitlement(EnterpriseAuthProvidersEntitlement) {
		t.Fatal("expected auth providers entitlement to be granted")
	}
	if !provider.hasEntitlement(EnterpriseModelProvidersEntitlement) {
		t.Fatal("expected model providers entitlement to be granted")
	}
	if !provider.hasEntitlement("ANY_RANDOM_ENTITLEMENT") {
		t.Fatal("expected any entitlement to be granted")
	}
}

func TestSetLicenseKeyBlocked(t *testing.T) {
	provider, err := NewProvider(context.Background(), nil, Config{
		LicenseKey: "via-config",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = provider.SetLicenseKey(context.Background(), "new-key")
	if err != ErrLicenseKeyViaConfiguration {
		t.Fatalf("expected ErrLicenseKeyViaConfiguration, got %v", err)
	}
}

func TestSetLicenseKeyAllowed(t *testing.T) {
	provider, err := NewProvider(context.Background(), nil, Config{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = provider.SetLicenseKey(context.Background(), "new-key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if provider.LicenseKey() != "new-key" {
		t.Fatalf("expected license key to be 'new-key', got %q", provider.LicenseKey())
	}
}

func TestRemoveLicenseKey(t *testing.T) {
	provider, err := NewProvider(context.Background(), nil, Config{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_ = provider.SetLicenseKey(context.Background(), "some-key")
	err = provider.RemoveLicenseKey(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if provider.LicenseKey() != "" {
		t.Fatalf("expected license key to be empty, got %q", provider.LicenseKey())
	}
}
