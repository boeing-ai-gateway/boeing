package imagepullsecret

import (
	"testing"
	"time"

	"github.com/boeing-ai-gateway/boeing/apiclient/types"
	"github.com/boeing-ai-gateway/boeing/pkg/imagepullsecrets"
	boeingv1 "github.com/boeing-ai-gateway/boeing/pkg/storage/apis/boeing.boeing.ai/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPopulateECRComputedStatusUsesBoeingServiceAccount(t *testing.T) {
	handler := New(nil, nil, "kubernetes", "boeing-mcp", "boeing", "boeing", nil, "https://issuer.example.com")
	secret := &boeingv1.ImagePullSecret{
		Spec: boeingv1.ImagePullSecretSpec{
			ECR: &types.ECRImagePullSecretConfig{},
		},
	}
	var status boeingv1.ImagePullSecretStatus

	handler.populateECRComputedStatus(secret, &status)

	if status.Subject != "system:serviceaccount:boeing:boeing" {
		t.Fatalf("unexpected ECR subject: %q", status.Subject)
	}
}

func TestShouldRefreshECRHonorsManualRequest(t *testing.T) {
	now := time.Date(2026, 5, 12, 12, 0, 0, 0, time.UTC)
	lastSuccess := metav1.NewTime(now.Add(-time.Hour))
	lastReconciled := metav1.NewTime(now.Add(-time.Minute))
	secret := &boeingv1.ImagePullSecret{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				imagepullsecrets.AnnotationECRRefreshRequestedAt: now.Format(time.RFC3339Nano),
			},
		},
		Spec: boeingv1.ImagePullSecretSpec{
			ECR: &types.ECRImagePullSecretConfig{
				RefreshSchedule: "0 0 * * *",
			},
		},
		Status: boeingv1.ImagePullSecretStatus{
			LastSuccessTime:    &lastSuccess,
			LastReconciledTime: &lastReconciled,
		},
	}
	handler := &Handler{now: func() time.Time { return now }}

	if !handler.shouldRefreshECR(secret, false) {
		t.Fatal("expected manual refresh request to force refresh")
	}

	reconciledAfterRequest := metav1.NewTime(now.Add(time.Minute))
	secret.Status.LastReconciledTime = &reconciledAfterRequest
	if handler.shouldRefreshECR(secret, false) {
		t.Fatal("did not expect already-observed manual refresh request to force refresh")
	}
}
