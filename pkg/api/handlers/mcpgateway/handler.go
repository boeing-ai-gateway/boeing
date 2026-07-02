package mcpgateway

import (
	"fmt"
	"maps"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/boeing-ai-gateway/boeing/apiclient/types"
	"github.com/boeing-ai-gateway/boeing/pkg/api"
	"github.com/boeing-ai-gateway/boeing/pkg/api/handlers"
	"github.com/boeing-ai-gateway/boeing/pkg/controller/handlers/systemmcpserver"
	gateway "github.com/boeing-ai-gateway/boeing/pkg/gateway/client"
	"github.com/boeing-ai-gateway/boeing/pkg/mcp"
	v1 "github.com/boeing-ai-gateway/boeing/pkg/storage/apis/boeing.boeing.ai/v1"
	"github.com/boeing-ai-gateway/boeing/pkg/system"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Handler struct {
	mcpSessionManager *mcp.SessionManager
	transport         http.RoundTripper
}

func NewHandler(mcpSessionManager *mcp.SessionManager) *Handler {
	return &Handler{
		mcpSessionManager: mcpSessionManager,
		transport:         otelhttp.NewTransport(http.DefaultTransport),
	}
}

func (h *Handler) Proxy(req api.Context) error {
	serverConfig, mcpURL, allowDifferentPaths, err := h.ensureServerIsDeployed(req)
	if err != nil {
		return fmt.Errorf("failed to ensure server is deployed: %v", err)
	}

	u, err := url.Parse(mcpURL)
	if err != nil {
		http.Error(req.ResponseWriter, err.Error(), http.StatusInternalServerError)
		return nil
	}

	(&httputil.ReverseProxy{
		Transport: h.transport,
		Director: func(r *http.Request) {
			r.Header.Set("X-Forwarded-Host", r.Host)
			scheme := "https"
			if strings.HasPrefix(r.Host, "localhost") || strings.HasPrefix(r.Host, "127.0.0.1") {
				scheme = "http"
			}
			r.Header.Set("X-Forwarded-Proto", scheme)

			r.Host = u.Host
			r.URL.Scheme = u.Scheme
			r.URL.Host = u.Host
			r.URL.Path = u.Path
			if rest := r.PathValue("rest"); allowDifferentPaths && rest != "" {
				if strings.HasPrefix(rest, "/") {
					r.URL.Path = rest
				} else {
					r.URL.Path = "/" + rest
				}
			}

			// Merge query parameters from the incoming request and the upstream URL.
			// Preserve all values; if a key exists in both, both values will be present.
			upstreamQuery := u.Query()
			origQuery := r.URL.Query()
			for k, vs := range origQuery {
				for _, v := range vs {
					upstreamQuery.Add(k, v)
				}
			}
			r.URL.RawQuery = upstreamQuery.Encode()

			for i := range serverConfig.PassthroughHeaderNames {
				if i < len(serverConfig.PassthroughHeaderValues) {
					r.Header.Set(serverConfig.PassthroughHeaderNames[i], serverConfig.PassthroughHeaderValues[i])
				}
			}
		},
	}).ServeHTTP(req.ResponseWriter, req.Request)

	return nil
}

func (h *Handler) ensureServerIsDeployed(req api.Context) (mcp.ServerConfig, string, bool, error) {
	mcpID := req.PathValue("mcp_id")

	if system.IsSystemMCPServerID(mcpID) {
		return h.ensureSystemServerIsDeployed(req, mcpID)
	}

	mcpID, mcpServer, mcpServerConfig, err := handlers.ServerForActionWithConnectID(req, mcpID)
	if err != nil {
		return mcp.ServerConfig{}, "", false, fmt.Errorf("failed to get mcp server config: %w", err)
	}
	if mcpServer.Spec.Template {
		return mcp.ServerConfig{}, "", false, apierrors.NewNotFound(schema.GroupResource{Group: "boeing.boeing.ai", Resource: "mcpserver"}, mcpID)
	}

	// Add-hoc authorization for boeingbot agents
	if mcpServerConfig.BoeingbotAgentName != "" {
		var agent v1.BoeingbotAgent
		if err = req.Get(&agent, mcpServerConfig.BoeingbotAgentName); err != nil {
			return mcp.ServerConfig{}, "", false, fmt.Errorf("failed to get boeingbot agent %q: %w", mcpServerConfig.BoeingbotAgentName, err)
		}
		if agent.Spec.UserID != req.User.GetUID() && (!req.UserCanImpersonate() || !req.UserIsAdmin()) {
			return mcp.ServerConfig{}, "", false, types.NewErrForbidden("user is not authorized to access boeingbot agent %q", mcpServerConfig.BoeingbotAgentName)
		}
	}

	url, err := h.mcpSessionManager.LaunchServer(req.Context(), mcpServerConfig)
	if err != nil {
		return mcp.ServerConfig{}, "", false, fmt.Errorf("failed to launch mcp server: %w", err)
	}

	return mcpServerConfig, url, mcpServerConfig.BoeingbotAgentName != "", nil
}

func (h *Handler) ensureSystemServerIsDeployed(req api.Context, mcpID string) (mcp.ServerConfig, string, bool, error) {
	var systemServer v1.SystemMCPServer
	if err := req.Get(&systemServer, mcpID); err != nil {
		return mcp.ServerConfig{}, "", false, fmt.Errorf("failed to get system MCP server %q: %w", mcpID, err)
	}

	if systemServer.Spec.Manifest.Enabled != nil && !*systemServer.Spec.Manifest.Enabled {
		return mcp.ServerConfig{}, "", false, apierrors.NewNotFound(schema.GroupResource{Group: "boeing.boeing.ai", Resource: "systemmcpserver"}, mcpID)
	}

	// Only look up credentials if the manifest has env vars without static values.
	// This avoids expensive credential lookups on the hot path for servers like
	// boeing-mcp-server where all env vars have static values.
	credEnv := make(map[string]string)
	var needsCredentials bool
	for _, env := range systemServer.Spec.Manifest.Env {
		if env.Value == "" {
			needsCredentials = true
			break
		}
	}

	if needsCredentials {
		credCtx := systemServer.Name
		creds, err := req.GatewayClient.ListCredentials(req.Context(), gateway.ListCredentialsOptions{
			CredentialContexts: []string{credCtx},
		})
		if err != nil {
			return mcp.ServerConfig{}, "", false, fmt.Errorf("failed to list credentials for system server: %w", err)
		}

		secretToolName := systemmcpserver.SecretInfoToolName(systemServer.Name)
		for _, cred := range creds {
			// Skip the secret info credential — those vars go to the shim only, not the MCP server.
			if cred.Name == secretToolName {
				continue
			}
			credDetail, err := req.GatewayClient.RevealCredential(req.Context(), []string{credCtx}, cred.Name)
			if err != nil {
				continue
			}
			maps.Copy(credEnv, credDetail.Secrets)
		}
	}

	// Retrieve the token exchange credential
	var secretsCred map[string]string
	tokenExchangeCred, err := req.GatewayClient.RevealCredential(req.Context(), []string{systemServer.Name}, systemmcpserver.SecretInfoToolName(systemServer.Name))
	if err == nil {
		secretsCred = tokenExchangeCred.Secrets
	}

	credEnv, err = mcp.MergeBoundCreds(req.Context(), req.LocalK8sClient, req.BoeingNamespace, systemServer.Spec.Manifest.Env, systemServer.Spec.Manifest.RemoteConfig, credEnv)
	if err != nil {
		return mcp.ServerConfig{}, "", false, fmt.Errorf("failed to resolve secret bindings: %w", err)
	}

	baseURL := strings.TrimSuffix(req.APIBaseURL, "/api")
	audiences := systemServer.ValidConnectURLs(baseURL)

	serverConfig, _, err := mcp.SystemServerToServerConfig(systemServer, audiences, baseURL, credEnv, secretsCred)
	if err != nil {
		return mcp.ServerConfig{}, "", false, fmt.Errorf("failed to convert system server to config: %w", err)
	}

	mcpURL, err := h.mcpSessionManager.LaunchServer(req.Context(), serverConfig)
	if err != nil {
		return mcp.ServerConfig{}, "", false, fmt.Errorf("failed to launch system MCP server: %w", err)
	}

	return serverConfig, mcpURL, false, nil
}
