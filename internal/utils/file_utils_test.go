package utils

import (
    "os"
    "testing"

    "github.com/docent-net/k8s-nas-maintenance/internal/kube"
)

func TestSaveAndLoadReplicas(t *testing.T) {
    filename := "test_replicas.json"
    defer os.Remove(filename)

    replicas := map[string]kube.Scalable{
        "deployment/test": &kube.DeploymentScaler{OriginalReplicas: 3},
    }

    err := SaveReplicasToFile(filename, replicas)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    loadedReplicas, err := LoadReplicasFromFile(filename)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    if len(loadedReplicas) != 1 {
        t.Errorf("expected 1 replica, got %d", len(loadedReplicas))
    }

    if loadedReplicas["deployment/test"].GetOriginalReplicas() != 3 {
        t.Errorf("expected original replicas to be 3, got %d", loadedReplicas["deployment/test"].GetOriginalReplicas())
    }
}
