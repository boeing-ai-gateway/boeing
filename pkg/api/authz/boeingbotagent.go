package authz

import (
	"net/http"

	"github.com/boeing-ai-gateway/nah/pkg/router"
	v1 "github.com/boeing-ai-gateway/boeing/pkg/storage/apis/boeing.boeing.ai/v1"
	"github.com/boeing-ai-gateway/boeing/pkg/system"
)

func (a *Authorizer) checkBoeingbotAgent(req *http.Request, resources *Resources, u User) (bool, error) {
	if resources.BoeingbotAgentID == "" {
		return true, nil
	}

	var agent v1.BoeingbotAgent
	if err := a.get(req.Context(), router.Key(system.DefaultNamespace, resources.BoeingbotAgentID), &agent); err != nil {
		return false, err
	}

	// If the user owns the workflow, then authorization is granted.
	if agent.Spec.UserID == u.GetUID() {
		resources.Authorizated.BoeingbotAgent = &agent
		return true, nil
	}

	// If the workflow belongs to a project and the user owns that project, authorization is granted.
	if resources.Authorizated.Project != nil && resources.Authorizated.Project.Spec.UserID == u.GetUID() && agent.Spec.ProjectID == resources.Authorizated.Project.Name {
		resources.Authorizated.BoeingbotAgent = &agent
		return true, nil
	}

	// If the user has impersonation + admin privileges, allow access to any agent.
	if u.CanImpersonate && u.IsAdmin {
		resources.Authorizated.BoeingbotAgent = &agent
		return true, nil
	}

	return false, nil
}
