package registry

import (
	"context"
	"testing"

	"github.com/boeing-ai-gateway/boeing/apiclient/types"
	v1 "github.com/boeing-ai-gateway/boeing/pkg/storage/apis/boeing.boeing.ai/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConvertMCPServerCatalogEntryToRegistryRemoteFixedURLHasRemote(t *testing.T) {
	entry := registryTestCatalogEntry(types.RemoteCatalogConfig{
		FixedURL: "https://api.example.com/mcp",
	})

	got, err := ConvertMCPServerCatalogEntryToRegistry(context.Background(), entry, "https://boeing.example.com", "com.example.boeing", newMimeFetcher())
	if err != nil {
		t.Fatal(err)
	}

	if got.Meta.Boeing != nil && got.Meta.Boeing.ConfigurationRequired {
		t.Fatalf("expected fixed URL entry to be directly connectable, got boeing meta %#v", got.Meta.Boeing)
	}
	if len(got.Server.Remotes) != 1 {
		t.Fatalf("remote count = %d, want 1", len(got.Server.Remotes))
	}
	if got.Server.Remotes[0].URL != "https://boeing.example.com/mcp-connect/remote-entry" {
		t.Fatalf("remote URL = %q, want mcp-connect URL", got.Server.Remotes[0].URL)
	}
}

func TestConvertMCPServerCatalogEntryToRegistryRemoteHostnameRequiresConfiguration(t *testing.T) {
	entry := registryTestCatalogEntry(types.RemoteCatalogConfig{
		Hostname: "api.example.com",
	})

	got, err := ConvertMCPServerCatalogEntryToRegistry(context.Background(), entry, "https://boeing.example.com", "com.example.boeing", newMimeFetcher())
	if err != nil {
		t.Fatal(err)
	}

	if got.Meta.Boeing == nil || !got.Meta.Boeing.ConfigurationRequired {
		t.Fatalf("expected hostname entry to require configuration, got boeing meta %#v", got.Meta.Boeing)
	}
	if len(got.Server.Remotes) != 0 {
		t.Fatalf("expected no remotes for hostname entry, got %#v", got.Server.Remotes)
	}
}

func TestConvertMCPServerCatalogEntryToRegistryRemoteURLTemplateRequiresConfiguration(t *testing.T) {
	entry := registryTestCatalogEntry(types.RemoteCatalogConfig{
		URLTemplate: "https://${WORKSPACE}.example.com/mcp",
	})

	got, err := ConvertMCPServerCatalogEntryToRegistry(context.Background(), entry, "https://boeing.example.com", "com.example.boeing", newMimeFetcher())
	if err != nil {
		t.Fatal(err)
	}

	if got.Meta.Boeing == nil || !got.Meta.Boeing.ConfigurationRequired {
		t.Fatalf("expected URL template entry to require configuration, got boeing meta %#v", got.Meta.Boeing)
	}
	if len(got.Server.Remotes) != 0 {
		t.Fatalf("expected no remotes for URL template entry, got %#v", got.Server.Remotes)
	}
}

func TestConvertMCPServerCatalogEntryToRegistryRemoteStaticOAuthRequiresConfigurationUntilConfigured(t *testing.T) {
	entry := registryTestCatalogEntry(types.RemoteCatalogConfig{
		FixedURL:            "https://api.example.com/mcp",
		StaticOAuthRequired: true,
	})

	got, err := ConvertMCPServerCatalogEntryToRegistry(context.Background(), entry, "https://boeing.example.com", "com.example.boeing", newMimeFetcher())
	if err != nil {
		t.Fatal(err)
	}

	if got.Meta.Boeing == nil || !got.Meta.Boeing.ConfigurationRequired {
		t.Fatalf("expected unconfigured static OAuth entry to require configuration, got boeing meta %#v", got.Meta.Boeing)
	}
	if len(got.Server.Remotes) != 0 {
		t.Fatalf("expected no remotes for unconfigured static OAuth entry, got %#v", got.Server.Remotes)
	}

	entry.Status.OAuthCredentialConfigured = true
	got, err = ConvertMCPServerCatalogEntryToRegistry(context.Background(), entry, "https://boeing.example.com", "com.example.boeing", newMimeFetcher())
	if err != nil {
		t.Fatal(err)
	}

	if got.Meta.Boeing != nil && got.Meta.Boeing.ConfigurationRequired {
		t.Fatalf("expected configured static OAuth entry to be directly connectable, got boeing meta %#v", got.Meta.Boeing)
	}
	if len(got.Server.Remotes) != 1 {
		t.Fatalf("remote count = %d, want 1", len(got.Server.Remotes))
	}
}

func TestConvertMCPServerToRegistryNeedsURLRequiresConfiguration(t *testing.T) {
	server := v1.MCPServer{
		ObjectMeta: metav1.ObjectMeta{
			Name: "ms1needsurl",
		},
		Spec: v1.MCPServerSpec{
			UserID:   "user-1",
			NeedsURL: true,
			Manifest: types.MCPServerManifest{
				Name:    "Needs URL",
				Runtime: types.RuntimeRemote,
				RemoteConfig: &types.RemoteRuntimeConfig{
					Hostname: "api.example.com",
				},
			},
		},
	}

	got, err := ConvertMCPServerToRegistry(context.Background(), server, nil, "https://boeing.example.com", server.Name, "com.example.boeing", "user-1", newMimeFetcher())
	if err != nil {
		t.Fatal(err)
	}

	if got.Meta.Boeing == nil || !got.Meta.Boeing.ConfigurationRequired {
		t.Fatalf("expected server needing URL to require configuration, got boeing meta %#v", got.Meta.Boeing)
	}
	if len(got.Server.Remotes) != 0 {
		t.Fatalf("expected no remotes for server needing URL, got %#v", got.Server.Remotes)
	}
}

func registryTestCatalogEntry(remoteConfig types.RemoteCatalogConfig) v1.MCPServerCatalogEntry {
	return v1.MCPServerCatalogEntry{
		ObjectMeta: metav1.ObjectMeta{
			Name: "remote-entry",
		},
		Spec: v1.MCPServerCatalogEntrySpec{
			Manifest: types.MCPServerCatalogEntryManifest{
				Name:        "Remote Entry",
				Description: "Remote entry",
				Runtime:     types.RuntimeRemote,
				RemoteConfig: &types.RemoteCatalogConfig{
					FixedURL:            remoteConfig.FixedURL,
					Hostname:            remoteConfig.Hostname,
					URLTemplate:         remoteConfig.URLTemplate,
					Headers:             remoteConfig.Headers,
					StaticOAuthRequired: remoteConfig.StaticOAuthRequired,
				},
			},
		},
	}
}
