package kube

import (
    "testing"

    appsv1 "k8s.io/api/apps/v1"
)

func TestDaemonSetScaler(t *testing.T) {
    daemonSet := &appsv1.DaemonSet{}
    scaler := &DaemonSetScaler{DaemonSet: daemonSet}

    replicas := int32(0)
    scaler.SetReplicas(replicas)
    if scaler.DaemonSet.Spec.UpdateStrategy.RollingUpdate == nil || scaler.DaemonSet.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable.IntVal != 0 {
        t.Errorf("expected MaxUnavailable to be 0")
    }

    replicas = 1
    scaler.SetReplicas(replicas)
    if scaler.DaemonSet.Spec.UpdateStrategy.RollingUpdate != nil {
        t.Errorf("expected RollingUpdate to be nil")
    }

    if scaler.GetReplicas() != 1 {
        t.Errorf("expected GetReplicas to return 1, got %d", scaler.GetReplicas())
    }

    if scaler.GetOriginalReplicas() != 1 {
        t.Errorf("expected GetOriginalReplicas to return 1, got %d", scaler.GetOriginalReplicas())
    }
}
