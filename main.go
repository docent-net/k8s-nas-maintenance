package main

import (
    "github.com/docent-net/k8s-nas-maintenance/cmd"
    "github.com/docent-net/k8s-nas-maintenance/internal/logging"
)

func main() {
    logging.InitLogger()
    defer logging.SyncLogger()

    cmd.Execute()
}
