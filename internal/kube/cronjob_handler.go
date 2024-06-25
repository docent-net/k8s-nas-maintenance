package kube

import (
    "context"
    "fmt"
    "sync"
    "time"

    "k8s.io/client-go/kubernetes"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func HandleCronJobsAndJobs(ctx context.Context, clientset *kubernetes.Clientset, namespace, storageClass string, wg *sync.WaitGroup) {
    defer wg.Done()

    // Find all CronJobs using the specific storage class
    namespaces := []string{namespace}
    if namespace == "" {
        nsList, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
        if err != nil {
            fmt.Printf("Error listing namespaces: %v\n", err)
            return
        }
        namespaces = []string{}
        for _, ns := range nsList.Items {
            namespaces = append(namespaces, ns.Name)
        }
    }

    for _, ns := range namespaces {
        cronJobs, err := clientset.BatchV1beta1().CronJobs(ns).List(ctx, metav1.ListOptions{})
        if err != nil {
            fmt.Printf("Error listing CronJobs: %v\n", err)
            return
        }

        for _, cronJob := range cronJobs.Items {
            // Check if the CronJob uses a PVC with the specified storage class
            for _, volume := range cronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes {
                if volume.PersistentVolumeClaim != nil {
                    pvc, err := clientset.CoreV1().PersistentVolumeClaims(ns).Get(ctx, volume.PersistentVolumeClaim.ClaimName, metav1.GetOptions{})
                    if err != nil {
                        fmt.Printf("Error getting PVC: %v\n", err)
                        continue
                    }
                    if *pvc.Spec.StorageClassName == storageClass {
                        // Suspend the CronJob
                        fmt.Printf("Suspending CronJob %s\n", cronJob.Name)
                        cronJob.Spec.Suspend = boolPtr(true)
                        _, err := clientset.BatchV1beta1().CronJobs(ns).Update(ctx, &cronJob, metav1.UpdateOptions{})
                        if err != nil {
                            fmt.Printf("Error suspending CronJob: %v\n", err)
                            continue
                        }

                        // Wait for all Jobs created by this CronJob to complete
                        fmt.Printf("Waiting for Jobs of CronJob %s to complete\n", cronJob.Name)
                        waitForJobsToComplete(ctx, clientset, ns, cronJob.Name)

                        // Resume the CronJob after maintenance
                        cronJob.Spec.Suspend = boolPtr(false)
                        _, err = clientset.BatchV1beta1().CronJobs(ns).Update(ctx, &cronJob, metav1.UpdateOptions{})
                        if err != nil {
                            fmt.Printf("Error resuming CronJob: %v\n", err)
                        }
                    }
                }
            }
        }
    }
}

func waitForJobsToComplete(ctx context.Context, clientset *kubernetes.Clientset, namespace, cronJobName string) {
    for {
        select {
        case <-ctx.Done():
            fmt.Printf("Context cancelled while waiting for jobs of CronJob %s to complete\n", cronJobName)
            return
        default:
            jobs, err := clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{
                LabelSelector: fmt.Sprintf("job-name=%s", cronJobName),
            })
            if err != nil {
                fmt.Printf("Error listing Jobs: %v\n", err)
                return
            }

            allCompleted := true
            for _, job := range jobs.Items {
                if job.Status.Succeeded == 0 {
                    allCompleted = false
                    break
                }
            }

            if allCompleted {
                fmt.Printf("All Jobs of CronJob %s have completed\n", cronJobName)
                return
            }

            time.Sleep(5 * time.Second)
        }
    }
}

func boolPtr(b bool) *bool {
    return &b
}
