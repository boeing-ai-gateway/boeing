package authz

import (
	"net/http"

	"github.com/boeing-ai-gateway/nah/pkg/router"
	v1 "github.com/boeing-ai-gateway/boeing/pkg/storage/apis/boeing.boeing.ai/v1"
	"github.com/boeing-ai-gateway/boeing/pkg/system"
)

func (a *Authorizer) checkSkill(req *http.Request, resources *Resources, u User) (bool, error) {
	if resources.SkillID == "" || u.IsAdmin || u.IsAuditor {
		return true, nil
	}

	var skill v1.Skill
	if err := a.get(req.Context(), router.Key(system.DefaultNamespace, resources.SkillID), &skill); err != nil {
		return false, err
	}

	return a.skillHelper.UserHasAccessToSkill(u, &skill)
}
