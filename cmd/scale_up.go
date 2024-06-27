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
)

var scaleUpCmd = &cobra.Command{
    Use:   "scale-up",
    Short: "Scale up Kubernetes workloads after NAS maintenance",
    Run:   runScaleUp,
}

func init() {
    rootCmd.AddCommand(scaleUpCmd)
}

func runScaleUp(cmd *cobra.Command, args []string) {
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

    workloadReplicas, err := utils.LoadReplicasFromFile(replicaFile)
    if err != nil {
        logging.Logger.Error("Error loading replicas from file", zap.Error(err))
        return
    }

    var wg sync.WaitGroup

    // Handle CronJobs and Jobs in a separate Goroutine
    wg.Add(1)
    go kube.HandleCronJobsAndJobs(ctx, clientset, namespace, storageClass, &wg)

    // Output or scale up the workloads
    for workload, scaler := range workloadReplicas {
        if dryRun {
            logging.Logger.Info("Would scale up", zap.String("workload", workload))
        } else {
            logging.Logger.Info("Scaling up", zap.String("workload", workload))
            kube.ScaleResource(clientset, scaler, namespace, scaler.GetOriginalReplicas())
        }
    }

    // Wait for CronJobs and Jobs to be handled
    wg.Wait()

    if dryRun {
        logging.Logger.Info("Dry run complete. No changes were made.")
    } else {
        logging.Logger.Info("Workloads have been scaled up.")
    }
}
