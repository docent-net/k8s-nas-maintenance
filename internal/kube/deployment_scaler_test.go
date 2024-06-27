package kube

import (
    "testing"

    appsv1 "k8s.io/api/apps/v1"
)

func TestDeploymentScaler(t *testing.T) {
    deployment := &appsv1.Deployment{
        Spec: appsv1.DeploymentSpec{
            Replicas: int32Ptr(3),
        },
    }
    scaler := &DeploymentScaler{Deployment: deployment}

    replicas := int32(3)
    scaler.SetReplicas(replicas)
    if *scaler.Deployment.Spec.Replicas != replicas {
        t.Errorf("expected replicas to be %d, got %d", replicas, *scaler.Deployment.Spec.Replicas)
    }

    if scaler.GetReplicas() != 0 {
        t.Errorf("expected GetReplicas to return 0, got %d", scaler.GetReplicas())
    }

    if scaler.GetOriginalReplicas() != 3 {
        t.Errorf("expected GetOriginalReplicas to return 3, got %d", scaler.GetOriginalReplicas())
    }
}

func int32Ptr(i int32) *int32 {
    return &i
}
