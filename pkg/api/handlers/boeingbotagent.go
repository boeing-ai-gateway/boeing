package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	nmcp "github.com/boeing-ai-gateway/boeingbot/pkg/mcp"
	"github.com/boeing-ai-gateway/boeing/apiclient/types"
	"github.com/boeing-ai-gateway/boeing/pkg/api"
	"github.com/boeing-ai-gateway/boeing/pkg/mcp"
	v1 "github.com/boeing-ai-gateway/boeing/pkg/storage/apis/boeing.boeing.ai/v1"
	"github.com/boeing-ai-gateway/boeing/pkg/system"
	"github.com/boeing-ai-gateway/boeing/pkg/wait"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type BoeingbotAgentHandler struct {
	sessionManager *mcp.SessionManager
	serverURL      string
	agentsEnabled  bool
}

func NewBoeingbotAgentHandler(sessionManager *mcp.SessionManager, serverURL string, agentsEnabled bool) *BoeingbotAgentHandler {
	return &BoeingbotAgentHandler{
		sessionManager: sessionManager,
		serverURL:      serverURL,
		agentsEnabled:  agentsEnabled,
	}
}

func (h *BoeingbotAgentHandler) ListAll(req api.Context) error {
	if !req.UserIsOwner() && !req.UserIsAdmin() && !req.UserIsAuditor() {
		return types.NewErrHTTP(http.StatusForbidden, "user is not authorized to list all boeingbot agents")
	}

	var agents v1.BoeingbotAgentList
	if err := req.List(&agents); err != nil {
		return err
	}

	items := make([]types.BoeingbotAgent, 0, len(agents.Items))
	for _, agent := range agents.Items {
		server, err := loadBoeingbotAgentMCPServer(req, agent)
		if err != nil {
			return err
		}
		items = append(items, h.convertBoeingbotAgent(agent, server))
	}
	return req.Write(types.BoeingbotAgentList{Items: items})
}

func (h *BoeingbotAgentHandler) List(req api.Context) error {
	var agents v1.BoeingbotAgentList
	if err := req.List(&agents, kclient.MatchingFields{
		"spec.projectID": req.PathValue("project_id"),
	}); err != nil {
		return err
	}

	items := make([]types.BoeingbotAgent, 0, len(agents.Items))
	for _, agent := range agents.Items {
		server, err := loadBoeingbotAgentMCPServer(req, agent)
		if err != nil {
			return err
		}
		items = append(items, h.convertBoeingbotAgent(agent, server))
	}
	return req.Write(types.BoeingbotAgentList{Items: items})
}

func (h *BoeingbotAgentHandler) Create(req api.Context) error {
	if !h.agentsEnabled {
		return types.NewErrHTTP(http.StatusForbidden, "Boeing Agent features are disabled")
	}

	var manifest types.BoeingbotAgentManifest
	if err := req.Read(&manifest); err != nil {
		return err
	}

	agent := v1.BoeingbotAgent{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: system.BoeingbotAgentPrefix,
			Namespace:    req.Namespace(),
		},
		Spec: v1.BoeingbotAgentSpec{
			BoeingbotAgentManifest: manifest,
			UserID:               req.User.GetUID(),
			ProjectID:            req.PathValue("project_id"),
		},
	}

	if err := req.Create(&agent); err != nil {
		return err
	}

	server, err := loadBoeingbotAgentMCPServer(req, agent)
	if err != nil {
		return err
	}
	return req.WriteCreated(h.convertBoeingbotAgent(agent, server))
}

func (h *BoeingbotAgentHandler) ByID(req api.Context) error {
	var agent v1.BoeingbotAgent
	if err := req.Get(&agent, req.PathValue("boeingbot_agent_id")); err != nil {
		return err
	}

	// Ensure that the agent belongs to the specified project
	if agent.Spec.ProjectID != req.PathValue("project_id") {
		return types.NewErrNotFound("boeingbot agent not found")
	}

	server, err := loadBoeingbotAgentMCPServer(req, agent)
	if err != nil {
		return err
	}
	return req.Write(h.convertBoeingbotAgent(agent, server))
}

func (h *BoeingbotAgentHandler) Update(req api.Context) error {
	var (
		id    = req.PathValue("boeingbot_agent_id")
		agent v1.BoeingbotAgent
	)

	if err := req.Get(&agent, id); err != nil {
		return err
	}

	// Ensure that the agent belongs to the specified project
	if agent.Spec.ProjectID != req.PathValue("project_id") {
		return types.NewErrNotFound("boeingbot agent not found")
	}

	var manifest types.BoeingbotAgentManifest
	if err := req.Read(&manifest); err != nil {
		return err
	}

	agent.Spec.BoeingbotAgentManifest = manifest
	if err := req.Update(&agent); err != nil {
		return err
	}

	server, err := loadBoeingbotAgentMCPServer(req, agent)
	if err != nil {
		return err
	}
	return req.Write(h.convertBoeingbotAgent(agent, server))
}

func (h *BoeingbotAgentHandler) Delete(req api.Context) error {
	var id = req.PathValue("boeingbot_agent_id")
	var agent v1.BoeingbotAgent
	if err := req.Get(&agent, id); err != nil {
		return err
	}

	// Ensure that the agent belongs to the specified project
	if agent.Spec.ProjectID != req.PathValue("project_id") {
		return types.NewErrNotFound("boeingbot agent not found")
	}

	return req.Delete(&v1.BoeingbotAgent{
		ObjectMeta: metav1.ObjectMeta{
			Name:      id,
			Namespace: req.Namespace(),
		},
	})
}

func (h *BoeingbotAgentHandler) Launch(req api.Context) error {
	var agent v1.BoeingbotAgent
	if err := req.Get(&agent, req.PathValue("boeingbot_agent_id")); err != nil {
		return err
	}

	if agent.Spec.ProjectID != req.PathValue("project_id") {
		return types.NewErrNotFound("boeingbot agent not found")
	}

	server := &v1.MCPServer{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: req.Namespace(),
			Name:      system.MCPServerPrefix + req.PathValue("boeingbot_agent_id"),
		},
	}

	ctx, cancel := context.WithTimeout(req.Context(), 15*time.Second)
	defer cancel()

	server, err := wait.For(ctx, req.Storage, server, func(srv *v1.MCPServer) (bool, error) {
		return srv.ResourceVersion != "", nil
	}, wait.Option{
		WaitForExists: true,
	})
	if err != nil {
		return fmt.Errorf("failed to load MCP server for agent %s: %w", agent.Name, err)
	}

	// Retry until credentials are available or the context deadline is reached.
	// On initial agent setup there is a race between MCPServer creation and credential
	// provisioning by the controller, so serverConfigForAction may transiently return a
	// "missing required config: BOEINGBOT_ENV_FILE" error before the credential exists.
	var serverConfig mcp.ServerConfig
	for {
		serverConfig, err = serverConfigForAction(req, *server)
		if err == nil {
			break
		}
		var errHTTP *types.ErrHTTP
		if !errors.As(err, &errHTTP) || errHTTP.Code != http.StatusBadRequest ||
			(!strings.Contains(errHTTP.Message, "BOEINGBOT_ENV_FILE") && !strings.Contains(errHTTP.Message, "BOEINGBOT_CONFIG_FILE")) {
			return err
		}
		select {
		case <-ctx.Done():
			return err
		case <-time.After(500 * time.Millisecond):
		}
	}

	if _, err = h.sessionManager.LaunchServer(req.Context(), serverConfig); err != nil {
		if errors.Is(err, mcp.ErrHealthCheckFailed) || errors.Is(err, mcp.ErrHealthCheckTimeout) {
			return types.NewErrHTTP(http.StatusServiceUnavailable, fmt.Sprintf("MCP server for agent %s is not healthy, check configuration for errors", agent.Name))
		}
		if errors.Is(err, nmcp.ErrNoResult) || strings.HasSuffix(err.Error(), nmcp.ErrNoResult.Error()) {
			return types.NewErrHTTP(http.StatusServiceUnavailable, fmt.Sprintf("No response from MCP server for agent %s, check configuration for errors", agent.Name))
		}
		if errors.Is(err, mcp.ErrInsufficientCapacity) {
			return types.NewErrHTTP(http.StatusServiceUnavailable, "Insufficient capacity to deploy MCP server for agent. Please contact your administrator.")
		}
		if nse := (*mcp.ErrNotSupportedByBackend)(nil); errors.As(err, &nse) {
			return types.NewErrHTTP(http.StatusBadRequest, nse.Error())
		}

		return fmt.Errorf("failed to launch MCP server for agent %s: %w", agent.Name, err)
	}

	return nil
}

func loadBoeingbotAgentMCPServer(req api.Context, agent v1.BoeingbotAgent) (*v1.MCPServer, error) {
	var server v1.MCPServer
	err := req.Get(&server, system.MCPServerPrefix+agent.Name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &server, nil
}

func (h *BoeingbotAgentHandler) convertBoeingbotAgent(agent v1.BoeingbotAgent, mcpServer *v1.MCPServer) types.BoeingbotAgent {
	out := types.BoeingbotAgent{
		Metadata:             MetadataFrom(&agent),
		BoeingbotAgentManifest: agent.Spec.BoeingbotAgentManifest,
		UserID:               agent.Spec.UserID,
		ProjectID:            agent.Spec.ProjectID,
		ConnectURL:           system.BoeingbotAgentConnectURL(h.serverURL, agent.Name),
	}
	if mcpServer != nil {
		out.NeedsURL = mcpServer.Spec.NeedsURL
		out.NeedsUpdate = mcpServer.Status.NeedsUpdate
		out.NeedsK8sUpdate = mcpServer.Status.NeedsK8sUpdate
		out.DeploymentStatus = mcpServer.Status.DeploymentStatus
	}
	return out
}
