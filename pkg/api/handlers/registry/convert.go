package registry

import (
	"context"
	"fmt"
	"strings"
	"time"

	boeingtypes "github.com/boeing-ai-gateway/boeing/apiclient/types"
	"github.com/boeing-ai-gateway/boeing/pkg/api/handlers"
	v1 "github.com/boeing-ai-gateway/boeing/pkg/storage/apis/boeing.boeing.ai/v1"
)

// ConvertMCPServerToRegistry converts an Boeing MCPServer to a Registry ServerResponse
// Uses the existing ConvertMCPServer function to ensure consistency with the rest of the codebase
func ConvertMCPServerToRegistry(
	ctx context.Context,
	server v1.MCPServer,
	credEnv map[string]string,
	serverURL string,
	slug string,
	reverseDNS string,
	userID string,
	mimeFetcher *mimeFetcher,
) (boeingtypes.RegistryServerResponse, error) {
	// Use existing conversion function to get types.MCPServer
	convertedServer := handlers.ConvertMCPServer(server, credEnv, serverURL, slug)

	// Generate registry server name
	displayName := convertedServer.MCPServerManifest.Name
	if displayName == "" {
		displayName = convertedServer.ID
	}

	if server.Spec.Alias != "" {
		displayName = server.Spec.Alias
	}

	registryName := FormatRegistryServerName(reverseDNS, slug)

	serverDetail := boeingtypes.RegistryServerDetail{
		Name:        registryName,
		Description: convertedServer.MCPServerManifest.ShortDescription,
		Title:       displayName,
		Version:     "latest",
		Schema:      "https://static.modelcontextprotocol.io/schemas/2025-09-29/server.schema.json",
		Meta: boeingtypes.RegistryServerMeta{
			PublisherProvided: &boeingtypes.RegistryPublisherProvidedMeta{
				GitHub: &boeingtypes.RegistryGitHubMeta{
					Readme: server.Spec.Manifest.Description,
				},
			},
		},
	}

	// Add icon if present
	if convertedServer.MCPServerManifest.Icon != "" {
		serverDetail.Icons = []boeingtypes.RegistryServerIcon{
			{
				Src:      convertedServer.MCPServerManifest.Icon,
				MimeType: mimeFetcher.guessMimeType(ctx, convertedServer.MCPServerManifest.Icon),
			},
		}
	}

	// Create metadata
	meta := boeingtypes.RegistryMeta{
		Official: boeingtypes.RegistryOfficialMeta{
			IsLatest:  true,
			CreatedAt: server.CreationTimestamp.Format(time.RFC3339),
			Status:    "active",
		},
	}

	// Determine if server should show connection URL
	isPersonalServer := convertedServer.UserID == userID && convertedServer.IsSingleUser()
	isMultiUserServer := !convertedServer.IsSingleUser()

	// For configured servers, add remote with mcp-connect URL
	// All Boeing servers are exposed as streamable-http remotes regardless of underlying runtime
	if isPersonalServer && convertedServer.Configured && !convertedServer.NeedsURL && convertedServer.ConnectURL != "" {
		// This is a personal server that is configured and ready to go.
		serverDetail.Remotes = []boeingtypes.RegistryServerRemote{
			{
				Type: "streamable-http",
				URL:  convertedServer.ConnectURL,
			},
		}
	} else if isMultiUserServer {
		// Multi-user servers are pre-configured by admins, so they always get a connection URL
		connectURL := fmt.Sprintf("%s/mcp-connect/%s", serverURL, server.Name)
		serverDetail.Remotes = []boeingtypes.RegistryServerRemote{
			{
				Type: "streamable-http",
				URL:  connectURL,
			},
		}
	} else {
		// Personal server that is not configured
		meta.Boeing = &boeingtypes.RegistryBoeingMeta{
			ConfigurationRequired: true,
			ConfigurationMessage:  "This server requires configuration. Please visit the Boeing UI to configure it.",
		}

		serverDetail.Meta.PublisherProvided.GitHub.Readme = fmt.Sprintf("> Note: This server requires configuration and cannot be installed directly from your client. Please visit [Boeing](%s) to to configure this server and obtain a connection URL.\n\n%s", serverURL, serverDetail.Meta.PublisherProvided.GitHub.Readme)
	}

	return boeingtypes.RegistryServerResponse{
		Server:        serverDetail,
		Meta:          meta,
		CreatedAtUnix: server.CreationTimestamp.Unix(),
	}, nil
}

// ConvertMCPServerCatalogEntryToRegistry converts a catalog entry to Registry format
func ConvertMCPServerCatalogEntryToRegistry(
	ctx context.Context,
	entry v1.MCPServerCatalogEntry,
	serverURL string,
	reverseDNS string,
	mimeFetcher *mimeFetcher,
) (boeingtypes.RegistryServerResponse, error) {
	manifest := entry.Spec.Manifest

	// Generate registry server name
	displayName := manifest.Name
	if displayName == "" {
		displayName = entry.Name
	}
	registryName := FormatRegistryServerName(reverseDNS, entry.Name)

	serverDetail := boeingtypes.RegistryServerDetail{
		Name:        registryName,
		Description: manifest.ShortDescription,
		Title:       displayName,
		Version:     "latest",
		Schema:      "https://static.modelcontextprotocol.io/schemas/2025-09-29/server.schema.json",
		Meta: boeingtypes.RegistryServerMeta{
			PublisherProvided: &boeingtypes.RegistryPublisherProvidedMeta{
				GitHub: &boeingtypes.RegistryGitHubMeta{
					Readme: entry.Spec.Manifest.Description,
				},
			},
		},
	}

	// Add icon if present
	if manifest.Icon != "" {
		serverDetail.Icons = []boeingtypes.RegistryServerIcon{
			{
				Src:      manifest.Icon,
				MimeType: mimeFetcher.guessMimeType(ctx, manifest.Icon),
			},
		}
	}

	// Add repository if present
	if manifest.RepoURL != "" {
		source := guessRepoSource(manifest.RepoURL)
		if source != "" {
			serverDetail.Repository = &boeingtypes.RegistryServerRepository{
				URL:    manifest.RepoURL,
				Source: source,
			}
		}
	}

	requiresConfiguration := catalogEntryRequiresConfiguration(entry)

	// Create metadata
	meta := boeingtypes.RegistryMeta{
		Official: boeingtypes.RegistryOfficialMeta{
			IsLatest:  true,
			CreatedAt: entry.CreationTimestamp.Format(time.RFC3339),
			Status:    "active",
		},
	}

	if requiresConfiguration {
		// Requires configuration - show configuration message
		meta.Boeing = &boeingtypes.RegistryBoeingMeta{
			ConfigurationRequired: true,
			ConfigurationMessage:  "This server needs to be configured before use. Please visit the Boeing UI to set it up.",
		}

		serverDetail.Meta.PublisherProvided.GitHub.Readme = fmt.Sprintf("> Note: This server requires configuration and cannot be installed directly from your client. Please visit [Boeing](%s) to to configure this server and obtain a connection URL.\n\n%s", serverURL, serverDetail.Meta.PublisherProvided.GitHub.Readme)
	} else {
		// No configuration required - provide connection URL
		serverDetail.Remotes = []boeingtypes.RegistryServerRemote{
			{
				Type: "streamable-http",
				URL:  fmt.Sprintf("%s/mcp-connect/%s", serverURL, entry.Name),
			},
		}
	}

	return boeingtypes.RegistryServerResponse{
		Server:        serverDetail,
		Meta:          meta,
		CreatedAtUnix: entry.CreationTimestamp.Unix(),
	}, nil
}

// Helper functions

func catalogEntryRequiresConfiguration(entry v1.MCPServerCatalogEntry) bool {
	manifest := entry.Spec.Manifest

	// Composite servers always require configuration in the UI before they can be used.
	if manifest.Runtime == boeingtypes.RuntimeComposite {
		return true
	}

	for _, env := range manifest.Env {
		// Required env values without a secret binding must be configured
		if env.Required && env.SecretBinding == nil {
			return true
		}
	}

	if manifest.Runtime == boeingtypes.RuntimeRemote && manifest.RemoteConfig != nil {
		if manifest.RemoteConfig.StaticOAuthRequired && !entry.Status.OAuthCredentialConfigured {
			return true
		}

		// Without a fixed URL, the user must supply a connection URL.
		if manifest.RemoteConfig.FixedURL == "" {
			return true
		}

		for _, header := range manifest.RemoteConfig.Headers {
			if header.Required && header.Value == "" && header.SecretBinding == nil {
				return true
			}
		}
	}

	return false
}

func guessRepoSource(repoURL string) string {
	lower := strings.ToLower(repoURL)
	if strings.Contains(lower, "github.com") {
		return "github"
	}
	if strings.Contains(lower, "gitlab.com") {
		return "gitlab"
	}
	if strings.Contains(lower, "bitbucket.org") {
		return "bitbucket"
	}
	return ""
}
