package utils

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/docent-net/k8s-nas-maintenance/internal/kube"
)

type ScalableWrapper struct {
    Type             string          `json:"type"`
    Object           json.RawMessage `json:"object"`
    OriginalReplicas int32           `json:"originalReplicas"`
}

func SaveReplicasToFile(filename string, replicas map[string]kube.Scalable) error {
    file, err := os.Create(filename)
    if err != nil {
        return fmt.Errorf("error creating file: %v", err)
    }
    defer file.Close()

    wrappers := make(map[string]ScalableWrapper)
    for k, v := range replicas {
        var typ string
        var obj json.RawMessage
        var originalReplicas int32
        switch scaler := v.(type) {
        case *kube.DeploymentScaler:
            typ = "deployment"
            obj, err = json.Marshal(scaler)
            originalReplicas = scaler.GetOriginalReplicas()
        case *kube.StatefulSetScaler:
            typ = "statefulset"
            obj, err = json.Marshal(scaler)
            originalReplicas = scaler.GetOriginalReplicas()
        case *kube.DaemonSetScaler:
            typ = "daemonset"
            obj, err = json.Marshal(scaler)
            originalReplicas = scaler.GetOriginalReplicas()
        }
        if err != nil {
            return fmt.Errorf("error marshaling scalable object: %v", err)
        }
        wrappers[k] = ScalableWrapper{Type: typ, Object: obj, OriginalReplicas: originalReplicas}
    }

    encoder := json.NewEncoder(file)
    err = encoder.Encode(wrappers)
    if err != nil {
        return fmt.Errorf("error encoding JSON: %v", err)
    }

    return nil
}

func LoadReplicasFromFile(filename string) (map[string]kube.Scalable, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, fmt.Errorf("error opening file: %v", err)
    }
    defer file.Close()

    wrappers := make(map[string]ScalableWrapper)
    decoder := json.NewDecoder(file)
    err = decoder.Decode(&wrappers)
    if err != nil {
        return nil, fmt.Errorf("error decoding JSON: %v", err)
    }

    replicas := make(map[string]kube.Scalable)
    for k, v := range wrappers {
        var obj kube.Scalable
        switch v.Type {
        case "deployment":
            obj = &kube.DeploymentScaler{}
        case "statefulset":
            obj = &kube.StatefulSetScaler{}
        case "daemonset":
            obj = &kube.DaemonSetScaler{}
        default:
            return nil, fmt.Errorf("unknown scalable type: %s", v.Type)
        }
        err = json.Unmarshal(v.Object, obj)
        if err != nil {
            return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
        }
        // Set the original replicas
        switch scaler := obj.(type) {
        case *kube.DeploymentScaler:
            scaler.OriginalReplicas = v.OriginalReplicas
        case *kube.StatefulSetScaler:
            scaler.OriginalReplicas = v.OriginalReplicas
        case *kube.DaemonSetScaler:
            scaler.OriginalReplicas = v.OriginalReplicas
        }
        replicas[k] = obj
    }

    return replicas, nil
}
