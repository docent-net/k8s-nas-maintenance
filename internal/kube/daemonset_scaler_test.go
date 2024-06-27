package kube

import (
    "context"
    "testing"

    appsv1 "k8s.io/api/apps/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes/fake"
)

func TestDaemonSetScaler(t *testing.T) {
    daemonSet := &appsv1.DaemonSet{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "test-daemonset",
            Namespace: "default",
        },
    }

    clientset := fake.NewSimpleClientset(daemonSet)
    scaler := &DaemonSetScaler{DaemonSet: daemonSet}

    ctx := context.TODO()

    // Test setting replicas to 0 (scaling down)
    scaler.SetReplicas(0)
    if replicas := scaler.GetReplicas(); replicas != 1 {
        t.Errorf("expected GetReplicas to return 1, got %d", replicas)
    }
    if err := scaler.Update(clientset, "default", ctx); err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    // Verify the daemonset was updated
    updatedDaemonSet, err := clientset.AppsV1().DaemonSets("default").Get(ctx, "test-daemonset", metav1.GetOptions{})
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if updatedDaemonSet.Spec.UpdateStrategy.RollingUpdate == nil || updatedDaemonSet.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.IntValue() != 0 {
        t.Errorf("expected MaxUnavailable to be 0")
    }

    // Test setting replicas back to the original value (scaling up)
    scaler.SetReplicas(1)
    if replicas := scaler.GetReplicas(); replicas != 1 {
        t.Errorf("expected GetReplicas to return 1, got %d", replicas)
    }
    if err := scaler.Update(clientset, "default", ctx); err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    // Verify the daemonset was updated
    updatedDaemonSet, err = clientset.AppsV1().DaemonSets("default").Get(ctx, "test-daemonset", metav1.GetOptions{})
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if updatedDaemonSet.Spec.UpdateStrategy.RollingUpdate != nil {
        t.Errorf("expected RollingUpdate to be reset")
    }
}
