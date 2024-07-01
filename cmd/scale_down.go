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
    "golang.org/x/time/rate"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/util/retry"
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
    dryRun, _ := cmd.Flags().GetBool("dry-run")

    kubeconfig := clientcmd.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
    config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
    if err != nil {
        logging.Logger.Error("Error loading kubeconfig", zap.Error(err))
        return
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        logging.Logger.Error("Error creating Kubernetes client", zap.Error(err))
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
    defer cancel()

    limiter := rate.NewLimiter(rate.Every(200*time.Millisecond), 1) // Adjust rate as necessary
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
                                logging.Logger.Debug("ownerRef Kind / Name", zap.String("ReplicaSet", ownerRef.Kind),
                                zap.String("ReplicaSet", ownerRef.Name))
                                switch ownerRef.Kind {
                                case "ReplicaSet":
                                    rs, err := clientset.AppsV1().ReplicaSets(ns).Get(ctx, ownerRef.Name, metav1.GetOptions{})
                                    if err != nil {
                                        logging.Logger.Error("Error getting ReplicaSet", zap.Error(err))
                                        continue
                                    }
                                    if len(rs.OwnerReferences) == 0 {
                                        logging.Logger.Error("ReplicaSet has no owner references; we don't support " +
                                        "standalone replicasets", zap.String("ReplicaSet", rs.Name))
                                        continue
                                    } else if len(rs.OwnerReferences) > 1 {
                                        logging.Logger.Error("ReplicaSet has more than 1 owner references; we don't support " +
                                        "this kind of RSes", zap.String("ReplicaSet", rs.Name))
                                        continue
                                    } else { // Deployment case
                                        rsRefOwner := rs.OwnerReferences[0]
                                        if rsRefOwner.Kind == "Deployment" {
                                            deployment, err := clientset.AppsV1().Deployments(ns).Get(ctx, rsRefOwner.Name, metav1.GetOptions{})
                                            if err != nil {
                                                logging.Logger.Error("Error getting Deployment", zap.Error(err))
                                                continue
                                            }
                                            scaler := &kube.DeploymentScaler{Deployment: deployment}
                                            logging.Logger.Info("Current Deployment replicas",
                                                zap.String("name", deployment.Name),
                                                zap.String("namespace", ns),
                                                zap.Int32("replicas", scaler.GetReplicas()),
                                            )
                                            if scaler.GetReplicas() == 0 {
                                                logging.Logger.Info("Skipping Deployment already at 0 replicas", zap.String("name", deployment.Name), zap.String("namespace", ns))
                                                continue
                                            }
                                            workloadReplicas["deployment/"+deployment.Name] = scaler
                                        } else {
                                            logging.Logger.Error("ReplicaSet owner is not a Deployment; we don't " +
                                            "this", zap.String("ownerName", rsRefOwner.Name), zap.String("ownerKind", rsRefOwner.Kind))
                                        continue
                                        }

                                    }
                                case "StatefulSet":
                                    statefulSet, err := clientset.AppsV1().StatefulSets(ns).Get(ctx, ownerRef.Name, metav1.GetOptions{})
                                    if err != nil {
                                        logging.Logger.Error("Error getting StatefulSet", zap.Error(err))
                                        continue
                                    }
                                    scaler := &kube.StatefulSetScaler{StatefulSet: statefulSet}
                                    logging.Logger.Info("Current StatefulSet replicas",
                                        zap.String("name", statefulSet.Name),
                                        zap.String("namespace", ns),
                                        zap.Int32("replicas", scaler.GetReplicas()),
                                    )
                                    if scaler.GetReplicas() == 0 {
                                        logging.Logger.Info("Skipping StatefulSet already at 0 replicas", zap.String("name", statefulSet.Name), zap.String("namespace", ns))
                                        continue
                                    }
                                    workloadReplicas["statefulset/"+statefulSet.Name] = scaler
                                case "DaemonSet":
                                    daemonSet, err := clientset.AppsV1().DaemonSets(ns).Get(ctx, ownerRef.Name, metav1.GetOptions{})
                                    if err != nil {
                                        logging.Logger.Error("Error getting DaemonSet", zap.Error(err))
                                        continue
                                    }
                                    scaler := &kube.DaemonSetScaler{DaemonSet: daemonSet}
                                    logging.Logger.Info("Current DaemonSet replicas",
                                        zap.String("name", daemonSet.Name),
                                        zap.String("namespace", ns),
                                        zap.Int32("replicas", scaler.GetReplicas()),
                                    )
                                    workloadReplicas["daemonset/"+daemonSet.Name] = scaler
                                case "Job":
                                    job, err := clientset.BatchV1().Jobs(ns).Get(ctx, ownerRef.Name, metav1.GetOptions{})
                                    if err != nil {
                                        logging.Logger.Error("Error getting Job", zap.Error(err))
                                        continue
                                    }
                                    logging.Logger.Info("Waiting for Job to complete", zap.String("jobName", job.Name))
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
            logging.Logger.Info("Would scale down", zap.String("workload", workload), zap.Int32("originalReplicas", scaler.GetOriginalReplicas()))
        } else {
            limiter.Wait(ctx)
            retry.OnError(retry.DefaultBackoff, func(error) bool { return true }, func() error {
                logging.Logger.Info("Scaling down", zap.String("workload", workload), zap.Int32("originalReplicas", scaler.GetOriginalReplicas()))
                scaler.SetReplicas(0)
                return scaler.Update(clientset, namespace, ctx)
            })
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
