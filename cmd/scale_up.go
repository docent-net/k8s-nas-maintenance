package cmd

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/spf13/cobra"
    "github.com/docent-net/k8s-nas-maintenance/internal/kube"
    "github.com/docent-net/k8s-nas-maintenance/internal/utils"
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
        fmt.Printf("Error creating in-cluster config: %v\n", err)
        return
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        fmt.Printf("Error creating Kubernetes client: %v\n", err)
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
    defer cancel()

    workloadReplicas, err := utils.LoadReplicasFromFile(replicaFile)
    if err != nil {
        fmt.Printf("Error loading replicas from file: %v\n", err)
        return
    }

    var wg sync.WaitGroup

    // Handle CronJobs and Jobs in a separate Goroutine
    wg.Add(1)
    go kube.HandleCronJobsAndJobs(ctx, clientset, namespace, storageClass, &wg)

    // Scale up the workloads
    for workload, scaler := range workloadReplicas {
        fmt.Printf("Scaling up %s\n", workload)
        kube.ScaleResource(clientset, scaler, namespace, scaler.GetOriginalReplicas())
    }

    // Wait for CronJobs and Jobs to be handled
    wg.Wait()

    fmt.Println("Workloads have been scaled up.")
}
