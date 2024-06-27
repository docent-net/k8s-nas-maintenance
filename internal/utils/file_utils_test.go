package utils

import (
    "os"
    "testing"

    appsv1 "k8s.io/api/apps/v1"
    "github.com/docent-net/k8s-nas-maintenance/internal/kube"
)

func int32Ptr(i int32) *int32 { return &i }

func TestSaveAndLoadReplicas(t *testing.T) {
    filename := "test_replicas.json"
    defer os.Remove(filename)

    deployment := &appsv1.Deployment{Spec: appsv1.DeploymentSpec{Replicas: int32Ptr(3)}}
    statefulSet := &appsv1.StatefulSet{Spec: appsv1.StatefulSetSpec{Replicas: int32Ptr(3)}}
    daemonSet := &appsv1.DaemonSet{}
    replicas := map[string]kube.Scalable{
        "deployment/test": &kube.DeploymentScaler{Deployment: deployment, OriginalReplicas: 3},
        "statefulset/test": &kube.StatefulSetScaler{StatefulSet: statefulSet, OriginalReplicas: 3},
        "daemonset/test": &kube.DaemonSetScaler{DaemonSet: daemonSet, OriginalReplicas: 1},
    }

    err := SaveReplicasToFile(filename, replicas)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    loadedReplicas, err := LoadReplicasFromFile(filename)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    if len(loadedReplicas) != 3 {
        t.Errorf("expected 3 replicas, got %d", len(loadedReplicas))
    }

    if loadedReplicas["deployment/test"].GetOriginalReplicas() != 3 {
        t.Errorf("expected original replicas to be 3, got %d", loadedReplicas["deployment/test"].GetOriginalReplicas())
    }

    if loadedReplicas["statefulset/test"].GetOriginalReplicas() != 3 {
        t.Errorf("expected original replicas to be 3, got %d", loadedReplicas["statefulset/test"].GetOriginalReplicas())
    }

    if loadedReplicas["daemonset/test"].GetOriginalReplicas() != 1 {
        t.Errorf("expected original replicas to be 1, got %d", loadedReplicas["daemonset/test"].GetOriginalReplicas())
    }
}
