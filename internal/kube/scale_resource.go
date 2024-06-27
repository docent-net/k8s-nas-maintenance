package kube

import (
    "context"
    "k8s.io/client-go/kubernetes"
)

func ScaleResourceWithContext(ctx context.Context, clientset kubernetes.Interface, scaler Scalable, namespace string, replicas int32) error {
    scaler.SetReplicas(replicas)
    return scaler.Update(clientset, namespace, ctx)
}
