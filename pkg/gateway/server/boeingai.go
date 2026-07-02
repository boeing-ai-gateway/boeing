package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/boeing-ai-gateway/boeing/pkg/api"
	"github.com/boeing-ai-gateway/boeing/pkg/gateway/types"
)

const (
	boeingAICredentialContext = "boeing-ai-provider"
	boeingAICredentialName    = "boeing-ai-provider"
	boeingAIAPIKeyField       = "BOEING_AI_MODEL_PROVIDER_API_KEY"
)

// getBoeingAICredentialStatus returns the status of the current user's Boeing AI credential.
func (s *Server) getBoeingAICredentialStatus(apiContext api.Context) error {
	userID := apiContext.User.GetUID()
	ctx := apiContext.Context()

	credContext := fmt.Sprintf("%s-%s", userID, boeingAICredentialContext)

	cred, err := apiContext.GatewayClient.RevealCredential(ctx, []string{credContext}, boeingAICredentialName)
	if err != nil {
		// Credential not found — not configured
		return apiContext.Write(map[string]any{
			"configured": false,
		})
	}

	// Credential exists — return masked token info
	token := cred.Secrets[boeingAIAPIKeyField]
	maskedToken := maskToken(token)

	return apiContext.Write(map[string]any{
		"configured":  true,
		"maskedToken": maskedToken,
	})
}

// saveBoeingAICredential saves the user's UDAL PAT as their Boeing AI credential.
func (s *Server) saveBoeingAICredential(apiContext api.Context) error {
	userID := apiContext.User.GetUID()
	ctx := apiContext.Context()

	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(apiContext.Request.Body).Decode(&req); err != nil {
		return writeError(apiContext, http.StatusBadRequest, "Invalid request body")
	}

	if req.Token == "" {
		return writeError(apiContext, http.StatusBadRequest, "Token is required")
	}

	// Store the credential
	credContext := fmt.Sprintf("%s-%s", userID, boeingAICredentialContext)

	cred := types.Credential{
		Context: credContext,
		Name:    boeingAICredentialName,
		Secrets: map[string]string{
			boeingAIAPIKeyField: req.Token,
		},
	}

	if err := apiContext.GatewayClient.UpsertCredential(ctx, cred); err != nil {
		return writeError(apiContext, http.StatusInternalServerError, fmt.Sprintf("Failed to save credential: %v", err))
	}

	return apiContext.Write(map[string]any{
		"success": true,
	})
}

// deleteBoeingAICredential removes the user's stored Boeing AI credential.
func (s *Server) deleteBoeingAICredential(apiContext api.Context) error {
	userID := apiContext.User.GetUID()
	ctx := apiContext.Context()

	credContext := fmt.Sprintf("%s-%s", userID, boeingAICredentialContext)

	if _, err := apiContext.GatewayClient.DeleteCredential(ctx, credContext, boeingAICredentialName); err != nil {
		return writeError(apiContext, http.StatusInternalServerError, fmt.Sprintf("Failed to delete credential: %v", err))
	}

	return apiContext.Write(map[string]any{
		"success": true,
	})
}

// maskToken returns a masked version of a token for display (e.g., "abc...xyz")
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "..." + token[len(token)-4:]
}

func writeError(apiContext api.Context, status int, message string) error {
	apiContext.ResponseWriter.Header().Set("Content-Type", "application/json")
	apiContext.ResponseWriter.WriteHeader(status)
	return json.NewEncoder(apiContext.ResponseWriter).Encode(map[string]any{
		"error": message,
	})
}

// loadPerUserBoeingAICredential checks if the model provider is Boeing AI and loads the user's
// stored UDAL PAT into credEnv so it gets injected via X-Boeing-* headers into the provider sidecar.
func (s *Server) loadPerUserBoeingAICredential(req api.Context, userID, modelProvider string, credEnv map[string]string) (map[string]string, bool) {
	// Check if this is the Boeing AI model provider
	if !isBoeingAIProvider(modelProvider) {
		return credEnv, false
	}

	ctx := req.Context()
	credContext := fmt.Sprintf("%s-%s", userID, boeingAICredentialContext)

	cred, err := req.GatewayClient.RevealCredential(ctx, []string{credContext}, boeingAICredentialName)
	if err != nil {
		// No per-user credential found — fall through to global key
		return credEnv, false
	}

	token, ok := cred.Secrets[boeingAIAPIKeyField]
	if !ok || token == "" {
		return credEnv, false
	}

	// Inject the per-user token into credEnv
	if credEnv == nil {
		credEnv = make(map[string]string)
	}
	credEnv[boeingAIAPIKeyField] = token

	return credEnv, true
}

// isBoeingAIProvider returns true if the model provider name matches the Boeing AI provider.
func isBoeingAIProvider(modelProvider string) bool {
	// Match by name patterns that indicate Boeing AI provider
	return modelProvider == "boeing-ai-model-provider" ||
		modelProvider == "boeing-ai" ||
		// Also match if the provider contains "boeing" in case of different naming
		containsBoeing(modelProvider)
}

func containsBoeing(s string) bool {
	for i := 0; i <= len(s)-6; i++ {
		if s[i:i+6] == "boeing" || s[i:i+6] == "Boeing" {
			return true
		}
	}
	return false
}
