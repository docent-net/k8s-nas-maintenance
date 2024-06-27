package kube

import (
    "context"
    "testing"
    "time"

    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/kubernetes/fake"
)

type MockScaler struct {
    replicas         int32
    originalReplicas int32
}

func (m *MockScaler) GetReplicas() int32 {
    return m.replicas
}

func (m *MockScaler) SetReplicas(replicas int32) {
    if replicas > 0 {
        m.originalReplicas = m.replicas
    }
    m.replicas = replicas
}

func (m *MockScaler) GetOriginalReplicas() int32 {
    return m.originalReplicas
}

func (m *MockScaler) Update(clientset kubernetes.Interface, namespace string, ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    case <-time.After(2 * time.Millisecond): // Simulate some processing time
        // Simulate an update operation
        return nil
    }
}

func TestScaleResourceWithContext(t *testing.T) {
    clientset := fake.NewSimpleClientset()
    scaler := &MockScaler{replicas: 3}

    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()

    err := ScaleResourceWithContext(ctx, clientset, scaler, "default", 1)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    if scaler.GetReplicas() != 1 {
        t.Errorf("expected replicas to be 1, got %d", scaler.GetReplicas())
    }

    // Test context timeout
    ctx, cancel = context.WithTimeout(context.Background(), 1*time.Nanosecond)
    defer cancel()

    time.Sleep(2 * time.Nanosecond) // Ensure the context has expired

    err = ScaleResourceWithContext(ctx, clientset, scaler, "default", 2)
    if err == nil {
        t.Errorf("expected an error due to context timeout")
    }
}
