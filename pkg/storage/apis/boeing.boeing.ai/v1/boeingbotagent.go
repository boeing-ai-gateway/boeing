package v1

import (
	"slices"

	"github.com/boeing-ai-gateway/nah/pkg/fields"
	"github.com/boeing-ai-gateway/boeing/apiclient/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	_ fields.Fields = (*BoeingbotAgent)(nil)
	_ DeleteRefs    = (*BoeingbotAgent)(nil)
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BoeingbotAgent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BoeingbotAgentSpec   `json:"spec,omitempty"`
	Status BoeingbotAgentStatus `json:"status,omitempty"`
}

func (in *BoeingbotAgent) Has(field string) (exists bool) {
	return slices.Contains(in.FieldNames(), field)
}

func (in *BoeingbotAgent) Get(field string) (value string) {
	switch field {
	case "spec.userID":
		return in.Spec.UserID
	case "spec.projectID":
		return in.Spec.ProjectID
	case "spec.projectV2ID":
		return in.Spec.ProjectV2ID
	}
	return ""
}

func (in *BoeingbotAgent) FieldNames() []string {
	return []string{"spec.userID", "spec.projectID", "spec.projectV2ID"}
}

func (in *BoeingbotAgent) DeleteRefs() []Ref {
	return []Ref{
		{
			ObjType: &Project{},
			Name:    in.Spec.ProjectID,
		},
	}
}

type BoeingbotAgentSpec struct {
	types.BoeingbotAgentManifest `json:",inline"`

	// UserID is the user that created this boeingbot workflow
	UserID string `json:"userID,omitempty"`

	// ProjectID is the project this workflow belongs to
	ProjectID string `json:"projectID,omitempty"`

	// ProjectV2ID is the project this workflow belongs to
	// Deprecated: use ProjectID instead.
	ProjectV2ID string `json:"projectV2ID,omitempty"`
}

type BoeingbotAgentStatus struct{}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type BoeingbotAgentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []BoeingbotAgent `json:"items"`
}
