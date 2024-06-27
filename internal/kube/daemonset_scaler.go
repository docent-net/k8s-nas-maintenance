package kube

import (
    "context"

    "k8s.io/client-go/kubernetes"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    appsv1 "k8s.io/api/apps/v1"
    "k8s.io/apimachinery/pkg/util/intstr"
    "github.com/docent-net/k8s-nas-maintenance/internal/logging"
    "go.uber.org/zap"
)

type DaemonSetScaler struct {
    DaemonSet        *appsv1.DaemonSet
    OriginalReplicas int32
}

func (d *DaemonSetScaler) GetReplicas() int32 {
    return 1 // DaemonSets do not use replicas directly
}

func (d *DaemonSetScaler) SetReplicas(replicas int32) {
    if d.OriginalReplicas == 0 {
        d.OriginalReplicas = 1
    }
    if replicas == 0 {
        d.DaemonSet.Spec.UpdateStrategy.RollingUpdate = &appsv1.RollingUpdateDaemonSet{MaxUnavailable: &intstr.IntOrString{IntVal: 0}}
    } else {
        d.DaemonSet.Spec.UpdateStrategy.RollingUpdate = nil // Reset to default
    }
}

func (d *DaemonSetScaler) GetOriginalReplicas() int32 {
    return d.OriginalReplicas
}

func (d *DaemonSetScaler) Update(clientset kubernetes.Interface, namespace string, ctx context.Context) error {
    logging.Logger.Info("Scaling down DaemonSet", zap.String("name", d.DaemonSet.Name), zap.String("namespace", namespace))
    _, err := clientset.AppsV1().DaemonSets(namespace).Update(ctx, d.DaemonSet, metav1.UpdateOptions{})
    if err != nil {
        logging.Logger.Error("Error updating DaemonSet", zap.String("name", d.DaemonSet.Name), zap.Error(err))
    }
    return err
}
