package kube

import (
    "context"
    "testing"

    appsv1 "k8s.io/api/apps/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes/fake"
)

func TestStatefulSetScaler(t *testing.T) {
    statefulSet := &appsv1.StatefulSet{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "test-statefulset",
            Namespace: "default",
        },
        Spec: appsv1.StatefulSetSpec{
            Replicas: int32Ptr(3),
        },
    }

    clientset := fake.NewSimpleClientset(statefulSet)
    scaler := &StatefulSetScaler{StatefulSet: statefulSet}

    ctx := context.TODO()

    // Test setting replicas to 0 (scaling down)
    scaler.SetReplicas(0)
    if replicas := scaler.GetReplicas(); replicas != 0 {
        t.Errorf("expected GetReplicas to return 0, got %d", replicas)
    }
    if err := scaler.Update(clientset, "default", ctx); err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    // Verify the statefulset was scaled down
    updatedStatefulSet, err := clientset.AppsV1().StatefulSets("default").Get(ctx, "test-statefulset", metav1.GetOptions{})
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if replicas := *updatedStatefulSet.Spec.Replicas; replicas != 0 {
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

    // Verify the statefulset was scaled up
    updatedStatefulSet, err = clientset.AppsV1().StatefulSets("default").Get(ctx, "test-statefulset", metav1.GetOptions{})
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if replicas := *updatedStatefulSet.Spec.Replicas; replicas != 3 {
        t.Errorf("expected replicas to be 3, got %d", replicas)
    }
}
