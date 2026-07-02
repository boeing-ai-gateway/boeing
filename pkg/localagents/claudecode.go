package localagents

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/boeing-ai-gateway/boeing/pkg/devicescan"
	"github.com/boeing-ai-gateway/boeing/pkg/localagents/assets"
)

const (
	ClaudeCodeAgentID     = "claude-code"
	claudeCodeDisplayName = "Claude Code"
)

type ClaudeCode struct {
	home string
}

func NewClaudeCode() ClaudeCode {
	return ClaudeCode{}
}

func (c ClaudeCode) ID() string {
	return ClaudeCodeAgentID
}

func (c ClaudeCode) DisplayName() string {
	return claudeCodeDisplayName
}

func (c ClaudeCode) Detect(ctx context.Context) DetectionResult {
	result := DetectionResult{
		AgentID:     c.ID(),
		DisplayName: c.DisplayName(),
		State:       DetectionMissing,
	}
	if err := ctx.Err(); err != nil {
		result.Reason = err.Error()
		return result
	}

	home, err := resolveHome("", c.home)
	if err != nil {
		result.Reason = err.Error()
		return result
	}

	presence := devicescan.DetectClaudeCodePresence(home)
	switch {
	case presence.BinaryPath != "":
		result.State = DetectionPresent
		result.Reason = "found claude binary at " + presence.BinaryPath
	case presence.ConfigPath != "":
		result.State = DetectionPresent
		result.Reason = "found Claude Code config at " + presence.ConfigPath
	case presence.InstallPath != "":
		result.State = DetectionPresent
		result.Reason = "found Claude Code install at " + presence.InstallPath
	default:
		result.Reason = "Claude Code was not detected"
	}

	return result
}

func (c ClaudeCode) InstallBootstrap(ctx context.Context, home string) (InstallResult, error) {
	if err := ctx.Err(); err != nil {
		return InstallResult{}, err
	}
	home, err := resolveHome(home, c.home)
	if err != nil {
		return InstallResult{}, err
	}

	rendered, err := assets.RenderAgentSkills(assets.ClaudeCodeTemplateData())
	if err != nil {
		return InstallResult{}, err
	}

	installed, err := installBootstrapAssets(claudeCodeSkillsRoot(home), rendered)
	if err != nil {
		return InstallResult{}, err
	}

	return InstallResult{
		AgentID:     c.ID(),
		DisplayName: c.DisplayName(),
		Installed:   installed,
		Message:     "Installed Boeing bootstrap skills for Claude Code",
	}, nil
}

func (c ClaudeCode) InstallSkill(ctx context.Context, home string, skill SkillArchive) (InstallResult, error) {
	if err := ctx.Err(); err != nil {
		return InstallResult{}, err
	}
	home, err := resolveHome(home, c.home)
	if err != nil {
		return InstallResult{}, err
	}
	name, installed, err := installSkillArchiveToRoot(claudeCodeSkillsRoot(home), skill)
	if err != nil {
		return InstallResult{}, err
	}

	return InstallResult{
		AgentID:     c.ID(),
		DisplayName: c.DisplayName(),
		Installed:   installed,
		Message:     fmt.Sprintf("Installed %s for Claude Code", name),
	}, nil
}

func claudeCodeSkillsRoot(home string) string {
	return filepath.Join(home, ".claude", "skills")
}
