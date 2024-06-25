package kube

import (
    "context"

    "k8s.io/client-go/kubernetes"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    appsv1 "k8s.io/api/apps/v1"
)

type DeploymentScaler struct {
    Deployment *appsv1.Deployment
    originalReplicas int32
}

func (d *DeploymentScaler) GetReplicas() int32 {
    return 0 // Scaling down
}

func (d *DeploymentScaler) SetReplicas(replicas int32) {
    if replicas > 0 {
        d.originalReplicas = replicas
    }
    d.Deployment.Spec.Replicas = &replicas
}

func (d *DeploymentScaler) GetOriginalReplicas() int32 {
    return d.originalReplicas
}

func (d *DeploymentScaler) Update(clientset *kubernetes.Clientset, namespace, name string) error {
    _, err := clientset.AppsV1().Deployments(namespace).Update(context.TODO(), d.Deployment, metav1.UpdateOptions{})
    return err
}
