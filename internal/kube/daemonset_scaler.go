package kube

import (
    "context"

    "k8s.io/client-go/kubernetes"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    appsv1 "k8s.io/api/apps/v1"
    "k8s.io/apimachinery/pkg/util/intstr"
)

type DaemonSetScaler struct {
    DaemonSet        *appsv1.DaemonSet
    OriginalReplicas int32
}

func (d *DaemonSetScaler) GetReplicas() int32 {
    return 1 // DaemonSets do not use replicas directly
}

func (d *DaemonSetScaler) SetReplicas(replicas int32) {
    if replicas == 0 {
        d.OriginalReplicas = 1
        d.DaemonSet.Spec.UpdateStrategy.RollingUpdate = &appsv1.RollingUpdateDaemonSet{MaxUnavailable: &intstr.IntOrString{IntVal: 0}}
    } else {
        d.DaemonSet.Spec.UpdateStrategy.RollingUpdate = nil // Reset to default
    }
}

func (d *DaemonSetScaler) GetOriginalReplicas() int32 {
    return d.OriginalReplicas
}

func (d *DaemonSetScaler) Update(clientset *kubernetes.Clientset, namespace, name string) error {
    _, err := clientset.AppsV1().DaemonSets(namespace).Update(context.TODO(), d.DaemonSet, metav1.UpdateOptions{})
    return err
}
