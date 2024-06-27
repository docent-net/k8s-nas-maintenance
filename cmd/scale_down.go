package cmd

import (
    "context"
    "sync"
    "time"

    "github.com/spf13/cobra"
    "github.com/docent-net/k8s-nas-maintenance/internal/kube"
    "github.com/docent-net/k8s-nas-maintenance/internal/logging"
    "github.com/docent-net/k8s-nas-maintenance/internal/utils"
    "go.uber.org/zap"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var scaleDownCmd = &cobra.Command{
    Use:   "scale-down",
    Short: "Scale down Kubernetes workloads for NAS maintenance",
    Run:   runScaleDown,
}

func init() {
    rootCmd.AddCommand(scaleDownCmd)
}

func runScaleDown(cmd *cobra.Command, args []string) {
    namespace, _ := cmd.Flags().GetString("namespace")
    storageClass, _ := cmd.Flags().GetString("storage-class")
    replicaFile, _ := cmd.Flags().GetString("replica-file")

    config, err := rest.InClusterConfig()
    if err != nil {
        logging.Logger.Error("Error creating in-cluster config", zap.Error(err))
        return
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        logging.Logger.Error("Error creating Kubernetes client", zap.Error(err))
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
    defer cancel()

    workloadReplicas := make(map[string]kube.Scalable)
    var wg sync.WaitGroup

    // Handle CronJobs and Jobs in a separate Goroutine
    wg.Add(1)
    go kube.HandleCronJobsAndJobs(ctx, clientset, namespace, storageClass, &wg)

    // Find all PVCs using the specific storage class
    namespaces := []string{namespace}
    if namespace == "" {
        nsList, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
        if err != nil {
            logging.Logger.Error("Error listing namespaces", zap.Error(err))
            return
        }
        namespaces = []string{}
        for _, ns := range nsList.Items {
            namespaces = append(namespaces, ns.Name)
        }
    }

    for _, ns := range namespaces {
        pvcs, err := clientset.CoreV1().PersistentVolumeClaims(ns).List(ctx, metav1.ListOptions{})
        if err != nil {
            logging.Logger.Error("Error listing PVCs", zap.Error(err))
            return
        }

        for _, pvc := range pvcs.Items {
            if *pvc.Spec.StorageClassName == storageClass {
                // Find all Pods using these PVCs
                pods, err := clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
                if err != nil {
                    logging.Logger.Error("Error listing Pods", zap.Error(err))
                    return
                }

                for _, pod := range pods.Items {
                    for _, volume := range pod.Spec.Volumes {
                        if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == pvc.Name {
                            ownerRefs := pod.OwnerReferences
                            for _, ownerRef := range ownerRefs {
                                switch ownerRef.Kind {
                                case "ReplicaSet":
                                    rs, err := clientset.AppsV1().ReplicaSets(ns).Get(ctx, ownerRef.Name, metav1.GetOptions{})
                                    if err != nil {
                                        logging.Logger.Error("Error getting ReplicaSet", zap.Error(err))
                                        continue
                                    }
                                    deploymentOwner := rs.OwnerReferences[0]
                                    if deploymentOwner.Kind == "Deployment" {
                                        deployment, err := clientset.AppsV1().Deployments(ns).Get(ctx, deploymentOwner.Name, metav1.GetOptions{})
                                        if err != nil {
                                            logging.Logger.Error("Error getting Deployment", zap.Error(err))
                                            continue
                                        }
                                        workloadReplicas["deployment/"+deployment.Name] = &kube.DeploymentScaler{Deployment: deployment}
                                    }
                                case "StatefulSet":
                                    statefulSet, err := clientset.AppsV1().StatefulSets(ns).Get(ctx, ownerRef.Name, metav1.GetOptions{})
                                    if err != nil {
                                        logging.Logger.Error("Error getting StatefulSet", zap.Error(err))
                                        continue
                                    }
                                    workloadReplicas["statefulset/"+statefulSet.Name] = &kube.StatefulSetScaler{StatefulSet: statefulSet}
                                case "DaemonSet":
                                    daemonSet, err := clientset.AppsV1().DaemonSets(ns).Get(ctx, ownerRef.Name, metav1.GetOptions{})
                                    if err != nil {
                                        logging.Logger.Error("Error getting DaemonSet", zap.Error(err))
                                        continue
                                    }
                                    workloadReplicas["daemonset/"+daemonSet.Name] = &kube.DaemonSetScaler{DaemonSet: daemonSet}
                                default:
                                    logging.Logger.Warn("Unknown resource kind", zap.String("kind", ownerRef.Kind), zap.String("name", ownerRef.Name))
                                }
                            }
                        }
                    }
                }
            }
        }
    }

    // Save original replicas to file
    if !dryRun {
        err = utils.SaveReplicasToFile(replicaFile, workloadReplicas)
        if err != nil {
            logging.Logger.Error("Error saving replicas to file", zap.Error(err))
            return
        }
    }

    // Output or scale down the workloads
    for workload, scaler := range workloadReplicas {
        if dryRun {
            logging.Logger.Info("Would scale down", zap.String("workload", workload))
        } else {
            logging.Logger.Info("Scaling down", zap.String("workload", workload))
            kube.ScaleResource(clientset, scaler, namespace, 0)
        }
    }

    // Wait for CronJobs and Jobs to be handled
    wg.Wait()

    if dryRun {
        logging.Logger.Info("Dry run complete. No changes were made.")
    } else {
        logging.Logger.Info("Workloads have been scaled down.")
    }
}
