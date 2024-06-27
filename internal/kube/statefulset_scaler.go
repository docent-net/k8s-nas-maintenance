package kube

import (
    "context"

    "k8s.io/client-go/kubernetes"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    appsv1 "k8s.io/api/apps/v1"
)

type StatefulSetScaler struct {
    StatefulSet      *appsv1.StatefulSet
    OriginalReplicas int32
}

func (s *StatefulSetScaler) GetReplicas() int32 {
    return 0 // Scaling down
}

func (s *StatefulSetScaler) SetReplicas(replicas int32) {
    if replicas > 0 {
        s.OriginalReplicas = *s.StatefulSet.Spec.Replicas
    }
    s.StatefulSet.Spec.Replicas = &replicas
}

func (s *StatefulSetScaler) GetOriginalReplicas() int32 {
    return s.OriginalReplicas
}

func (s *StatefulSetScaler) Update(clientset *kubernetes.Clientset, namespace, name string) error {
    _, err := clientset.AppsV1().StatefulSets(namespace).Update(context.TODO(), s.StatefulSet, metav1.UpdateOptions{})
    return err
}
