package kube

import (
    "testing"

    appsv1 "k8s.io/api/apps/v1"
)

func TestStatefulSetScaler(t *testing.T) {
    statefulSet := &appsv1.StatefulSet{}
    scaler := &StatefulSetScaler{StatefulSet: statefulSet}

    replicas := int32(3)
    scaler.SetReplicas(replicas)
    if *scaler.StatefulSet.Spec.Replicas != replicas {
        t.Errorf("expected replicas to be %d, got %d", replicas, *scaler.StatefulSet.Spec.Replicas)
    }

    if scaler.GetReplicas() != 0 {
        t.Errorf("expected GetReplicas to return 0, got %d", scaler.GetReplicas())
    }

    if scaler.GetOriginalReplicas() != replicas {
        t.Errorf("expected GetOriginalReplicas to return %d, got %d", replicas, scaler.GetOriginalReplicas())
    }
}
