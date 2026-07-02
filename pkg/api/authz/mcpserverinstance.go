package authz

import (
	"net/http"

	"github.com/boeing-ai-gateway/nah/pkg/router"
	v1 "github.com/boeing-ai-gateway/boeing/pkg/storage/apis/boeing.boeing.ai/v1"
	"github.com/boeing-ai-gateway/boeing/pkg/system"
)

func (a *Authorizer) checkMCPServerInstance(req *http.Request, resources *Resources, u User) (bool, error) {
	if resources.MCPServerInstanceID == "" {
		return true, nil
	}

	var mcpServerInstance v1.MCPServerInstance
	if err := a.get(req.Context(), router.Key(system.DefaultNamespace, resources.MCPServerInstanceID), &mcpServerInstance); err != nil {
		return false, err
	}

	return mcpServerInstance.Spec.UserID == u.GetUID(), nil
}
