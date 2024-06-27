package kube

import (
    "context"
    "fmt"
    "sync"
    "time"

    "k8s.io/client-go/kubernetes"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "go.uber.org/zap"
    "github.com/docent-net/k8s-nas-maintenance/internal/logging"
)

func HandleCronJobsAndJobs(ctx context.Context, clientset *kubernetes.Clientset, namespace, storageClass string, wg *sync.WaitGroup) {
    defer wg.Done()

    // Handle v1 CronJobs (stable in Kubernetes 1.21+)
    cronJobs, err := clientset.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{})
    if err != nil {
        logging.Logger.Error("Error listing CronJobs (v1)", zap.Error(err))
        return
    }

    for _, cronJob := range cronJobs.Items {
        if cronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes != nil {
            for _, volume := range cronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes {
                if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == storageClass {
                    // Pause the CronJob
                    cronJob.Spec.Suspend = boolPtr(true)
                    _, err := clientset.BatchV1().CronJobs(namespace).Update(ctx, &cronJob, metav1.UpdateOptions{})
                    if err != nil {
                        logging.Logger.Error("Error updating CronJob (v1)", zap.String("name", cronJob.Name), zap.Error(err))
                        continue
                    }
                    logging.Logger.Info("Paused CronJob (v1)", zap.String("name", cronJob.Name))

                    // Wait for all Jobs from this CronJob to complete
                    waitForJobsCompletion(ctx, clientset, cronJob.Name, namespace)
                }
            }
        }
    }

    // Handle standalone Jobs
    jobs, err := clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
    if err != nil {
        logging.Logger.Error("Error listing Jobs", zap.Error(err))
        return
    }

    for _, job := range jobs.Items {
        if job.Spec.Template.Spec.Volumes != nil {
            for _, volume := range job.Spec.Template.Spec.Volumes {
                if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == storageClass {
                    // Wait for the Job to complete
                    waitForSingleJobCompletion(ctx, clientset, job.Name, namespace)
                }
            }
        }
    }
}

func waitForJobsCompletion(ctx context.Context, clientset *kubernetes.Clientset, cronJobName, namespace string) {
    for {
        jobs, err := clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{
            LabelSelector: fmt.Sprintf("cronjob-name=%s", cronJobName),
        })
        if err != nil {
            logging.Logger.Error("Error listing Jobs", zap.Error(err))
            return
        }

        allCompleted := true
        for _, job := range jobs.Items {
            if job.Status.Active > 0 || job.Status.Failed > 0 {
                allCompleted = false
                break
            }
        }

        if allCompleted {
            logging.Logger.Info("All jobs completed for CronJob", zap.String("cronJobName", cronJobName))
            return
        }

        time.Sleep(5 * time.Second) // Wait before checking again
    }
}

func waitForSingleJobCompletion(ctx context.Context, clientset *kubernetes.Clientset, jobName, namespace string) {
    for {
        job, err := clientset.BatchV1().Jobs(namespace).Get(ctx, jobName, metav1.GetOptions{})
        if err != nil {
            logging.Logger.Error("Error getting Job", zap.String("name", jobName), zap.Error(err))
            return
        }

        logging.Logger.Info("Job stat1us", zap.String("jobName", jobName), zap.Int32("active", job.Status.Active), zap.Int32("failed", job.Status.Failed), zap.Int32("succeeded", job.Status.Succeeded))

        if job.Status.Active == 0 && job.Status.Failed == 0 {
            logging.Logger.Info("Job completed", zap.String("jobName", jobName))
            return
        }

        time.Sleep(5 * time.Second) // Wait before checking again
    }
}

func boolPtr(b bool) *bool {
    return &b
}

// New function to resume CronJobs
func ResumeCronJobs(ctx context.Context, clientset *kubernetes.Clientset, namespace, storageClass string, wg *sync.WaitGroup) {
    defer wg.Done()

    cronJobs, err := clientset.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{})
    if err != nil {
        logging.Logger.Error("Error listing CronJobs (v1)", zap.Error(err))
        return
    }

    for _, cronJob := range cronJobs.Items {
        if cronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes != nil {
            for _, volume := range cronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes {
                if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == storageClass {
                    // Resume the CronJob
                    cronJob.Spec.Suspend = boolPtr(false)
                    _, err := clientset.BatchV1().CronJobs(namespace).Update(ctx, &cronJob, metav1.UpdateOptions{})
                    if err != nil {
                        logging.Logger.Error("Error updating CronJob (v1)", zap.String("name", cronJob.Name), zap.Error(err))
                        continue
                    }
                    logging.Logger.Info("Resumed CronJob (v1)", zap.String("name", cronJob.Name))
                }
            }
        }
    }
}
