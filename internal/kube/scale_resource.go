package kube

import (
    "fmt"

    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/util/retry"
)

func ScaleResource(clientset *kubernetes.Clientset, scaler Scalable, namespace string, replicas int32) {
    retry.RetryOnConflict(retry.DefaultRetry, func() error {
        scaler.SetReplicas(replicas)
        err := scaler.Update(clientset, namespace, "")
        if err != nil {
            fmt.Printf("Error scaling resource: %v\n", err)
        }
        return err
    })
}
