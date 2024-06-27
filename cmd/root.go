package cmd

import (
    "os"

    "github.com/spf13/cobra"
)

var (
    dryRun bool
)

var rootCmd = &cobra.Command{
    Use:   "nas-maintenance",
    Short: "A tool for handling Kubernetes workloads during NAS maintenance",
}

// Execute executes the root command.
func Execute() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

func init() {
    rootCmd.PersistentFlags().StringP("namespace", "n", "", "Kubernetes namespace (default is all namespaces)")
    rootCmd.PersistentFlags().StringP("storage-class", "s", "", "Storage class name")
    rootCmd.PersistentFlags().StringP("replica-file", "r", "replicas.json", "File to store original replicas")
    rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Perform a dry run without making any changes")
    rootCmd.MarkPersistentFlagRequired("storage-class")
}
