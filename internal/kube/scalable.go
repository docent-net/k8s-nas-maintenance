package kube

import (
    "context"
    "k8s.io/client-go/kubernetes"
)

type Scalable interface {
    GetReplicas() int32
    SetReplicas(replicas int32)
    GetOriginalReplicas() int32
    Update(clientset kubernetes.Interface, namespace string, ctx context.Context) error
}
