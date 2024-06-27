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
    "k8s.io/client-go/util/retry"
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
    workloadReplicas, err := utils.LoadReplicasFromFile(replicaFile)
    if err != nil {
        logging.Logger.Error("Error loading replicas from file", zap.Error(err))
        return
    }

    var wg sync.WaitGroup

    // Resume CronJobs
    wg.Add(1)
    go kube.ResumeCronJobs(ctx, clientset, namespace, storageClass, &wg)

    // Scale up other workloads
    for workload, scaler := range workloadReplicas {
        wg.Add(1)
        go func(workload string, scaler kube.Scalable) {
            defer wg.Done()
            limiter.Wait(ctx)
            retry.OnError(retry.DefaultBackoff, func(error) bool { return true }, func() error {
                logging.Logger.Info("Scaling up", zap.String("workload", workload))
                originalReplicas := scaler.GetOriginalReplicas()
                return kube.ScaleResourceWithContext(ctx, clientset, scaler, namespace, originalReplicas)
            })
        }(workload, scaler)
    }

    wg.Wait()

    if dryRun {
        logging.Logger.Info("Dry run complete. No changes were made.")
    } else {
        logging.Logger.Info("Workloads have been scaled up.")
    }
}
