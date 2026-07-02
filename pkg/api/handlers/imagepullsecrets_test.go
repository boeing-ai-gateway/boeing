package handlers

import (
	"testing"

	"github.com/boeing-ai-gateway/boeing/apiclient/types"
	v1 "github.com/boeing-ai-gateway/boeing/pkg/storage/apis/boeing.boeing.ai/v1"
)

func TestImagePullSecretSpecFromInputDefaultsECRIssuerURL(t *testing.T) {
	handler := &ImagePullSecretHandler{issuerURL: "https://issuer.example.com"}

	spec, err := handler.specFromInput(types.ImagePullSecretManifest{
		Type: types.ImagePullSecretTypeECR,
		ECR: &types.ECRImagePullSecretConfig{
			RoleARN: "arn:aws:iam::123456789012:role/boeing-ecr",
			Region:  "us-east-1",
		},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.ECR == nil {
		t.Fatal("expected ECR spec")
	}
	if spec.ECR.IssuerURL != "https://issuer.example.com" {
		t.Fatalf("expected default issuer URL to be stored, got %q", spec.ECR.IssuerURL)
	}

	spec, err = handler.specFromInput(types.ImagePullSecretManifest{
		Type: types.ImagePullSecretTypeECR,
		ECR: &types.ECRImagePullSecretConfig{
			RoleARN:   "arn:aws:iam::123456789012:role/boeing-ecr",
			Region:    "us-east-1",
			IssuerURL: "https://custom-issuer.example.com/",
		},
	}, &v1.ImagePullSecret{Spec: v1.ImagePullSecretSpec{Type: types.ImagePullSecretTypeECR}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.ECR.IssuerURL != "https://custom-issuer.example.com" {
		t.Fatalf("expected explicit issuer URL to win, got %q", spec.ECR.IssuerURL)
	}
}

func TestImagePullSecretSpecFromInputRequiresECRIssuerURLWhenDefaultMissing(t *testing.T) {
	handler := &ImagePullSecretHandler{}

	_, err := handler.specFromInput(types.ImagePullSecretManifest{
		Type: types.ImagePullSecretTypeECR,
		ECR: &types.ECRImagePullSecretConfig{
			RoleARN: "arn:aws:iam::123456789012:role/boeing-ecr",
			Region:  "us-east-1",
		},
	}, nil)
	if err == nil {
		t.Fatal("expected issuerURL error")
	}

	spec, err := handler.specFromInput(types.ImagePullSecretManifest{
		Type: types.ImagePullSecretTypeECR,
		ECR: &types.ECRImagePullSecretConfig{
			RoleARN:   "arn:aws:iam::123456789012:role/boeing-ecr",
			Region:    "us-east-1",
			IssuerURL: "https://custom-issuer.example.com",
		},
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.ECR.IssuerURL != "https://custom-issuer.example.com" {
		t.Fatalf("expected explicit issuer URL, got %q", spec.ECR.IssuerURL)
	}
}

func TestImagePullSecretCapabilityReportsIssuerDiscoveryFailure(t *testing.T) {
	handler := &ImagePullSecretHandler{
		mcpRuntimeBackend: "kubernetes",
		issuerError:       "discovery failed",
	}

	capability := handler.convertCapability()
	if !capability.Available {
		t.Fatal("expected image pull secrets to remain available")
	}
	if capability.Reason == "" {
		t.Fatal("expected issuer discovery reason")
	}
}
