package kube

import (
    "testing"
)

func TestScaleResource(t *testing.T) {
    // Create a mock clientset and a scalable resource for testing
    clientset := &kubernetes.Clientset{}
    scaler := &MockScalable{}

    namespace := "default"
    replicas := int32(3)

    // Call ScaleResource with the mock clientset and scaler
    ScaleResource(clientset, scaler, namespace, replicas)

    // Verify that the replicas were set correctly
    if scaler.Replicas != replicas {
        t.Errorf("expected replicas to be %d, got %d", replicas, scaler.Replicas)
    }
}
