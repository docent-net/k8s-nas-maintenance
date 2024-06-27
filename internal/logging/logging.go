package logging

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func InitLogger() {
    config := zap.NewProductionConfig()
    config.EncoderConfig.TimeKey = "timestamp"
    config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
    config.OutputPaths = []string{"stdout"}

    var err error
    Logger, err = config.Build()
    if err != nil {
        panic(err)
    }
}

func SyncLogger() {
    _ = Logger.Sync() // flushes buffer, if any
}
