package types

// BoeingbotAgent represents a boeingbot workflow in the API
type BoeingbotAgent struct {
	Metadata
	BoeingbotAgentManifest
	UserID           string `json:"userID,omitempty"`
	ProjectID        string `json:"projectID,omitempty"`
	ProjectV2ID      string `json:"projectV2ID,omitempty"`
	ConnectURL       string `json:"connectURL,omitempty"`
	NeedsUpdate      bool   `json:"needsUpdate,omitempty"`
	NeedsK8sUpdate   bool   `json:"needsK8sUpdate,omitempty"`
	NeedsURL         bool   `json:"needsURL,omitempty"`
	DeploymentStatus string `json:"deploymentStatus,omitempty"`
}

// BoeingbotAgentManifest contains the user-editable fields for a boeingbot workflow
type BoeingbotAgentManifest struct {
	DisplayName  string `json:"displayName,omitempty"`
	Description  string `json:"description,omitempty"`
	DefaultAgent string `json:"defaultAgent,omitempty"`
}

// BoeingbotAgentList is a list of boeingbot workflows
type BoeingbotAgentList List[BoeingbotAgent]
