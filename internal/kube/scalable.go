package kube

import (
    "k8s.io/client-go/kubernetes"
)

type Scalable interface {
    GetReplicas() int32
    SetReplicas(int32)
    GetOriginalReplicas() int32
    Update(clientset *kubernetes.Clientset, namespace, name string) error
}
