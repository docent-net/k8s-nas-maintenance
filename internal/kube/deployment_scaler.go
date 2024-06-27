package kube

import (
    "context"

    "k8s.io/client-go/kubernetes"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    appsv1 "k8s.io/api/apps/v1"
    "github.com/docent-net/k8s-nas-maintenance/internal/logging"
    "go.uber.org/zap"
)

type DeploymentScaler struct {
    Deployment       *appsv1.Deployment
    OriginalReplicas int32
}

func (d *DeploymentScaler) GetReplicas() int32 {
    return *d.Deployment.Spec.Replicas
}

func (d *DeploymentScaler) SetReplicas(replicas int32) {
    if d.OriginalReplicas == 0 && *d.Deployment.Spec.Replicas != 0 {
        d.OriginalReplicas = *d.Deployment.Spec.Replicas
        logging.Logger.Info("Set original replicas for Deployment",
            zap.String("name", d.Deployment.Name),
            zap.Int32("originalReplicas", d.OriginalReplicas),
        )
    }
    logging.Logger.Info("Setting replicas for Deployment",
        zap.String("name", d.Deployment.Name),
        zap.Int32("newReplicas", replicas),
    )
    d.Deployment.Spec.Replicas = &replicas
}

func (d *DeploymentScaler) GetOriginalReplicas() int32 {
    return d.OriginalReplicas
}

func (d *DeploymentScaler) Update(clientset kubernetes.Interface, namespace string, ctx context.Context) error {
    logging.Logger.Info("Updating Deployment replicas",
        zap.String("name", d.Deployment.Name),
        zap.String("namespace", namespace),
        zap.Int32("currentReplicas", *d.Deployment.Spec.Replicas),
        zap.Int32("originalReplicas", d.OriginalReplicas),
    )

    _, err := clientset.AppsV1().Deployments(namespace).Update(ctx, d.Deployment, metav1.UpdateOptions{})
    if err != nil {
        logging.Logger.Error("Error updating Deployment",
            zap.String("name", d.Deployment.Name),
            zap.Error(err),
        )
    }
    return err
}
