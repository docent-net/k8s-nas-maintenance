package kube

import (
    "context"

    "k8s.io/client-go/kubernetes"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    appsv1 "k8s.io/api/apps/v1"
    "github.com/docent-net/k8s-nas-maintenance/internal/logging"
    "go.uber.org/zap"
)

type StatefulSetScaler struct {
    StatefulSet      *appsv1.StatefulSet
    OriginalReplicas int32
}

func (s *StatefulSetScaler) GetReplicas() int32 {
    return *s.StatefulSet.Spec.Replicas
}

func (s *StatefulSetScaler) SetReplicas(replicas int32) {
    if s.OriginalReplicas == 0 {
        s.OriginalReplicas = *s.StatefulSet.Spec.Replicas
    }
    s.StatefulSet.Spec.Replicas = &replicas
}

func (s *StatefulSetScaler) GetOriginalReplicas() int32 {
    return s.OriginalReplicas
}

func (s *StatefulSetScaler) Update(clientset kubernetes.Interface, namespace string, ctx context.Context) error {
    if *s.StatefulSet.Spec.Replicas == 0 {
        logging.Logger.Info("Skipping scale-down for StatefulSet already at 0 replicas", zap.String("name", s.StatefulSet.Name), zap.String("namespace", namespace))
        return nil
    }
    logging.Logger.Info("Scaling down StatefulSet", zap.String("name", s.StatefulSet.Name), zap.String("namespace", namespace), zap.Int32("currentReplicas", *s.StatefulSet.Spec.Replicas), zap.Int32("newReplicas", 0))
    _, err := clientset.AppsV1().StatefulSets(namespace).Update(ctx, s.StatefulSet, metav1.UpdateOptions{})
    if err != nil {
        logging.Logger.Error("Error updating StatefulSet", zap.String("name", s.StatefulSet.Name), zap.Error(err))
    }
    return err
}
