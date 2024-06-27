package kube

import (
    "context"
    "testing"

    appsv1 "k8s.io/api/apps/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes/fake"
)

func TestDeploymentScaler(t *testing.T) {
    deployment := &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "test-deployment",
            Namespace: "default",
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: int32Ptr(3),
        },
    }

    clientset := fake.NewSimpleClientset(deployment)
    scaler := &DeploymentScaler{Deployment: deployment}

    ctx := context.TODO()

    // Test setting replicas to 0 (scaling down)
    scaler.SetReplicas(0)
    if replicas := scaler.GetReplicas(); replicas != 0 {
        t.Errorf("expected GetReplicas to return 0, got %d", replicas)
    }
    if err := scaler.Update(clientset, "default", ctx); err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    // Verify the deployment was scaled down
    updatedDeployment, err := clientset.AppsV1().Deployments("default").Get(ctx, "test-deployment", metav1.GetOptions{})
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if replicas := *updatedDeployment.Spec.Replicas; replicas != 0 {
        t.Errorf("expected replicas to be 0, got %d", replicas)
    }

    // Test setting replicas back to the original value (scaling up)
    scaler.SetReplicas(3)
    if replicas := scaler.GetReplicas(); replicas != 3 {
        t.Errorf("expected GetReplicas to return 3, got %d", replicas)
    }
    if err := scaler.Update(clientset, "default", ctx); err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    // Verify the deployment was scaled up
    updatedDeployment, err = clientset.AppsV1().Deployments("default").Get(ctx, "test-deployment", metav1.GetOptions{})
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if replicas := *updatedDeployment.Spec.Replicas; replicas != 3 {
        t.Errorf("expected replicas to be 3, got %d", replicas)
    }
}
