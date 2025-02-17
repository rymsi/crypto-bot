package main

import (
	"ingestor-service/internal/config"
	"ingestor-service/internal/ingestor"
	"ingestor-service/internal/kafka"
	"ingestor-service/internal/websocket"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// Setup

	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Encoding = "console"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	logger, _ := loggerConfig.Build()

	defer logger.Sync()
	sugar := logger.Sugar()

	cfg, err := config.LoadIngestorConfig(sugar)
	if err != nil {
		sugar.Errorw("Failed to load config", "error", err)
		return
	}

	wsClient := websocket.NewClient(cfg.CoinbaseWSURL, cfg, sugar)
	kafkaProducer, err := kafka.NewProducer(cfg.KafkaBrokers, "btc_usd", sugar)
	if err != nil {
		sugar.Errorw("Failed to create kafka producer", "error", err)
		return
	}

	// Start Service
	sugar.Infow("Starting ingestor service...")

	ingestorService := ingestor.NewService(wsClient, kafkaProducer, sugar)
	err = ingestorService.Start()
	if err != nil {
		sugar.Errorw("Failed to start ingestor service", "error", err)
		return
	}

	defer func() {
		err = ingestorService.Stop()
		if err != nil {
			sugar.Errorw("Failed to stop ingestor service", "error", err)
		}
	}()

	// Wait for SIGINT or SIGTERM or timer to run out
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	sugar.Infow("Received SIGINT or SIGTERM, stopping ingestor service...")
	signal.Stop(sigChan)
}
