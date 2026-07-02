package license

import (
	"context"
	"errors"
	"sync"

	"github.com/boeing-ai-gateway/boeing/logger"
	"github.com/boeing-ai-gateway/boeing/pkg/gateway/client"
)

const (
	// LicenseKeyPropertyKey is the database property key used to persist the license key.
	LicenseKeyPropertyKey = "boeing-license-key"

	// LicenseMachineIDPropertyKey is the database property key used to persist the machine fingerprint.
	LicenseMachineIDPropertyKey = "boeing-license-machine-id"

	// EnterpriseAuthProvidersEntitlement is required to enable enterprise auth providers.
	EnterpriseAuthProvidersEntitlement = "BOEING_ENTERPRISE_AUTH_PROVIDERS"

	// EnterpriseModelProvidersEntitlement is required to enable enterprise model providers.
	EnterpriseModelProvidersEntitlement = "BOEING_ENTERPRISE_MODEL_PROVIDERS"
)

var (
	// ErrNotConfigured indicates license validation was requested without enough configuration.
	ErrNotConfigured = errors.New("license provider is not configured")

	// ErrLicenseKeyViaConfiguration indicates the license key is managed by startup configuration.
	ErrLicenseKeyViaConfiguration = errors.New("license key is configured at startup")

	// ErrInvalidLicense indicates the provided license key could not be validated.
	ErrInvalidLicense = errors.New("license key is invalid")

	log = logger.Package()
)

// Config contains license settings (retained for interface compatibility).
type Config struct {
	LicenseKey string `usage:"License key for this installation (unused — all features unlocked)"`
}

// entitlementCode is a local replacement for the keygen EntitlementCode type.
type entitlementCode string

// Provider is a no-op license provider that always reports a valid license
// with all entitlements granted. No external network calls are made.
type Provider struct {
	lock                       sync.RWMutex
	entitlements               map[entitlementCode]struct{}
	licenseKey                 string
	gatewayClient              *client.Client
	licenseKeyViaConfiguration bool
}

// NewProvider creates a license provider that always grants all entitlements.
// No external services are contacted.
func NewProvider(_ context.Context, gatewayClient *client.Client, config Config) (*Provider, error) {
	log.Infof("license provider: all features unlocked (no external license validation)")

	// Pre-populate with all known entitlements so nothing is ever gated.
	entitlements := map[entitlementCode]struct{}{
		entitlementCode(EnterpriseAuthProvidersEntitlement):  {},
		entitlementCode(EnterpriseModelProvidersEntitlement): {},
	}

	return &Provider{
		entitlements:               entitlements,
		licenseKey:                 config.LicenseKey,
		gatewayClient:              gatewayClient,
		licenseKeyViaConfiguration: config.LicenseKey != "",
	}, nil
}

func (p *Provider) LicenseKey() string {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.licenseKey
}

func (p *Provider) LicenseKeyViaConfiguration() bool {
	p.lock.RLock()
	defer p.lock.RUnlock()
	return p.licenseKeyViaConfiguration
}

func (p *Provider) SetLicenseKey(_ context.Context, licenseKey string) error {
	if p.LicenseKeyViaConfiguration() {
		return ErrLicenseKeyViaConfiguration
	}
	p.lock.Lock()
	p.licenseKey = licenseKey
	p.lock.Unlock()
	return nil
}

func (p *Provider) RemoveLicenseKey(_ context.Context) error {
	if p.LicenseKeyViaConfiguration() {
		return ErrLicenseKeyViaConfiguration
	}
	p.lock.Lock()
	p.licenseKey = ""
	p.lock.Unlock()
	return nil
}

// HasValidLicense always returns true — all features are unlocked.
func (p *Provider) HasValidLicense() bool {
	return true
}

// Entitlements returns all granted entitlements.
func (p *Provider) Entitlements() []string {
	p.lock.RLock()
	defer p.lock.RUnlock()

	entitlements := make([]string, 0, len(p.entitlements))
	for entitlement := range p.entitlements {
		entitlements = append(entitlements, string(entitlement))
	}
	return entitlements
}

// hasEntitlement always returns true — nothing is gated.
func (p *Provider) hasEntitlement(_ string) bool {
	return true
}
