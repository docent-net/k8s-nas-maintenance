package kube

import (
    "testing"

    "k8s.io/client-go/kubernetes"
)

type MockScalable struct {
    Replicas int32
}

func (m *MockScalable) GetReplicas() int32                { return 0 }
func (m *MockScalable) SetReplicas(replicas int32)        { m.Replicas = replicas }
func (m *MockScalable) GetOriginalReplicas() int32        { return m.Replicas }
func (m *MockScalable) Update(clientset *kubernetes.Clientset, namespace, name string) error { return nil }

func TestScaleResource(t *testing.T) {
    clientset := &kubernetes.Clientset{}
    scaler := &MockScalable{}

    namespace := "default"
    replicas := int32(3)

    ScaleResource(clientset, scaler, namespace, replicas)

    if scaler.Replicas != replicas {
        t.Errorf("expected replicas to be %d, got %d", replicas, scaler.Replicas)
    }
}
