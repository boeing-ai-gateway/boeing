package controller

import (
	"context"
	"testing"

	"github.com/boeing-ai-gateway/boeing/pkg/serviceaccounts"
	"github.com/boeing-ai-gateway/boeing/pkg/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeNetworkPolicyProviderInstaller struct {
	installed       *networkPolicyProviderInstallSpec
	uninstallCalled bool
	uninstallNS     string
}

func (f *fakeNetworkPolicyProviderInstaller) InstallOrUpgrade(_ context.Context, spec networkPolicyProviderInstallSpec) error {
	copied := spec
	copied.Values = cloneMap(spec.Values)
	f.installed = &copied
	return nil
}

func (f *fakeNetworkPolicyProviderInstaller) Uninstall(releaseNamespace string) error {
	f.uninstallCalled = true
	f.uninstallNS = releaseNamespace
	return nil
}

func newNetworkPolicyProviderController(t *testing.T, installer networkPolicyProviderInstaller) *Controller {
	t.Helper()

	return &Controller{
		services: &services.Services{
			MCPRuntimeBackend:                    "kubernetes",
			MCPServerNamespace:                   "boeing-mcp",
			MCPClusterDomain:                     "cluster.local",
			MCPNetworkPolicyEnabled:              true,
			MCPNetworkPolicyProviderChartRepo:    "https://charts.example.com",
			MCPNetworkPolicyProviderChartName:    "network-policy-provider",
			MCPNetworkPolicyProviderChartVersion: "1.2.3",
			ServiceName:                          "boeing",
			ServiceNamespace:                     "boeing-system",
			ServiceAccountName:                   "boeing-gateway",
			StorageListenPort:                    8443,
		},
		providerInstaller: installer,
	}
}

func TestEnsureNetworkPolicyProviderInstallsChartRelease(t *testing.T) {
	ctx := t.Context()
	installer := &fakeNetworkPolicyProviderInstaller{}
	controller := newNetworkPolicyProviderController(t, installer)

	require.NoError(t, controller.reconcileNetworkPolicyProvider(ctx))
	require.NotNil(t, installer.installed)
	assert.Equal(t, networkPolicyProviderReleaseName, installer.installed.ReleaseName)
	assert.Equal(t, "boeing-system", installer.installed.ReleaseNamespace)
	assert.Equal(t, "https://charts.example.com", installer.installed.ChartRepoURL)
	assert.Equal(t, "network-policy-provider", installer.installed.ChartName)
	assert.Equal(t, "1.2.3", installer.installed.ChartVersion)
	assert.Equal(t, "boeing-mcp", installer.installed.Values["mcpRuntimeNamespace"])
	assert.Equal(t, "https://boeing.boeing-system.svc.cluster.local:8443", installer.installed.Values["boeingStorageURL"])
	assert.Equal(t, serviceaccounts.NetworkPolicySecretName, installer.installed.Values["secretName"])
	assert.Equal(t, "/var/run/secrets/boeing-network-policy-provider/apiKey", installer.installed.Values["boeingStorageTokenFile"])
	boeingValues := installer.installed.Values["boeing"].(map[string]any)
	serviceAccountValues := boeingValues["serviceAccount"].(map[string]any)
	assert.Equal(t, "boeing-gateway", serviceAccountValues["name"])
	assert.Equal(t, "boeing-system", serviceAccountValues["namespace"])
}

func TestEnsureNetworkPolicyProviderMergesValuesBlob(t *testing.T) {
	ctx := t.Context()
	installer := &fakeNetworkPolicyProviderInstaller{}
	controller := newNetworkPolicyProviderController(t, installer)
	controller.services.MCPNetworkPolicyProviderValues = `
mcpRuntimeNamespace: custom-runtime
extraFlag: true
`

	require.NoError(t, controller.reconcileNetworkPolicyProvider(ctx))
	require.NotNil(t, installer.installed)
	assert.Equal(t, "custom-runtime", installer.installed.Values["mcpRuntimeNamespace"])
	assert.Equal(t, true, installer.installed.Values["extraFlag"])
	assert.Equal(t, "https://boeing.boeing-system.svc.cluster.local:8443", installer.installed.Values["boeingStorageURL"])
}

func TestEnsureNetworkPolicyProviderUninstallsWhenDisabled(t *testing.T) {
	ctx := t.Context()
	installer := &fakeNetworkPolicyProviderInstaller{}
	controller := newNetworkPolicyProviderController(t, installer)
	controller.services.MCPNetworkPolicyEnabled = false

	require.NoError(t, controller.reconcileNetworkPolicyProvider(ctx))
	assert.True(t, installer.uninstallCalled)
	assert.Equal(t, "boeing-system", installer.uninstallNS)
	assert.Nil(t, installer.installed)
}

func TestEnsureNetworkPolicyProviderSkipsUninstallOutsideKubernetes(t *testing.T) {
	ctx := t.Context()
	installer := &fakeNetworkPolicyProviderInstaller{}
	controller := newNetworkPolicyProviderController(t, installer)
	controller.services.MCPRuntimeBackend = "docker"
	controller.services.MCPNetworkPolicyEnabled = false

	require.NoError(t, controller.reconcileNetworkPolicyProvider(ctx))
	assert.False(t, installer.uninstallCalled)
	assert.Nil(t, installer.installed)
}

func TestEnsureNetworkPolicyProviderRequiresStorageSettings(t *testing.T) {
	ctx := t.Context()
	installer := &fakeNetworkPolicyProviderInstaller{}
	controller := newNetworkPolicyProviderController(t, installer)
	controller.services.ServiceName = ""

	err := controller.reconcileNetworkPolicyProvider(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service name")
}

func TestEnsureNetworkPolicyProviderRequiresServiceAccountName(t *testing.T) {
	ctx := t.Context()
	installer := &fakeNetworkPolicyProviderInstaller{}
	controller := newNetworkPolicyProviderController(t, installer)
	controller.services.ServiceAccountName = ""

	err := controller.reconcileNetworkPolicyProvider(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service account name")
}

func TestEnsureNetworkPolicyProviderUsesConfiguredClusterDomain(t *testing.T) {
	ctx := t.Context()
	installer := &fakeNetworkPolicyProviderInstaller{}
	controller := newNetworkPolicyProviderController(t, installer)
	controller.services.MCPClusterDomain = "example.internal"

	require.NoError(t, controller.reconcileNetworkPolicyProvider(ctx))
	require.NotNil(t, installer.installed)
	assert.Equal(t, "https://boeing.boeing-system.svc.example.internal:8443", installer.installed.Values["boeingStorageURL"])
}
