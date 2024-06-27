package utils

import (
    "encoding/json"
    "io/ioutil"
    "os"

    "github.com/docent-net/k8s-nas-maintenance/internal/kube"
    "github.com/docent-net/k8s-nas-maintenance/internal/logging"
    "go.uber.org/zap"
)

func SaveReplicasToFile(filePath string, workloadReplicas map[string]kube.Scalable) error {
    replicasMap := make(map[string]int32)
    for workload, scaler := range workloadReplicas {
        replicasMap[workload] = scaler.GetOriginalReplicas()
        logging.Logger.Info("Saving original replicas",
            zap.String("workload", workload),
            zap.Int32("originalReplicas", scaler.GetOriginalReplicas()),
        )
    }

    data, err := json.Marshal(replicasMap)
    if err != nil {
        logging.Logger.Error("Error marshalling replicas map", zap.Error(err))
        return err
    }

    err = ioutil.WriteFile(filePath, data, os.ModePerm)
    if err != nil {
        logging.Logger.Error("Error writing replicas to file", zap.Error(err))
        return err
    }

    return nil
}
