package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"go.uber.org/zap"
)

// RegisterCoralogixSink registers a new sink, and returns the output path
func RegisterCoralogixSink(key string) string {
	coralogixSink := NewCoralogixZapSinkFactory(key, "testApp", "testSubSys")
	err := zap.RegisterSink(CORALOGIX_SINK_SCHEME, coralogixSink)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%s:", CORALOGIX_SINK_SCHEME)
}

func CreateLogger() *zap.Logger {
	coralogixPrivateKey := os.Getenv("CORALOGIX_PRIVATE_KEY")
	coralogixOutputPath := RegisterCoralogixSink(coralogixPrivateKey)

	loggerConfig := zap.NewProductionConfig()
	loggerConfig.OutputPaths = append(loggerConfig.OutputPaths, coralogixOutputPath)

	logger, err := loggerConfig.Build()
	if err != nil {
		log.Fatal(err)
	}
	return logger
}

func main() {
	logger := CreateLogger()
	defer logger.Sync()

	logger.Info("failed to fetch URL",
		// Structured context as strongly typed Field values.
		zap.String("hello", "world"),
		zap.Int("attempt", 3),
		zap.Duration("backoff", time.Second),
	)
}
