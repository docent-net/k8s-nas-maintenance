package utils

import (
    "encoding/json"
    "fmt"
    "os"

    "github.com/docent-net/k8s-nas-maintenance/internal/kube"
)

func SaveReplicasToFile(filename string, replicas map[string]kube.Scalable) error {
    file, err := os.Create(filename)
    if err != nil {
        return fmt.Errorf("error creating file: %v", err)
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    err = encoder.Encode(replicas)
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

    var replicas map[string]kube.Scalable
    decoder := json.NewDecoder(file)
    err = decoder.Decode(&replicas)
    if err != nil {
        return nil, fmt.Errorf("error decoding JSON: %v", err)
    }

    return replicas, nil
}
