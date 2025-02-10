package main

import (
	"ingestor-service/internal/config"
	"ingestor-service/internal/ingestor"
	"ingestor-service/internal/kafka"
	"ingestor-service/internal/websocket"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	// Setup
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar := logger.Sugar()

	cfg, err := config.LoadIngestorConfig(sugar)
	if err != nil {
		sugar.Errorw("Failed to load config", "error", err)
		return
	}

	wsClient := websocket.NewClient(cfg.CoinbaseWSURL, sugar)
	kafkaProducer, err := kafka.NewProducer(cfg.KafkaBrokers, "btc-usd", sugar)
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
		sugar.Infow("Stopping ingestor service...")
		err = ingestorService.Stop()
		if err != nil {
			sugar.Errorw("Failed to stop ingestor service", "error", err)
		}
		sugar.Infow("Ingestor service stopped")
	}()

	// Wait for SIGINT or SIGTERM or timer to run out
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	timer := time.NewTimer(300 * time.Second)

	select {
	case <-sigChan:
		sugar.Infow("Received SIGINT or SIGTERM, stopping ingestor service...")
		signal.Stop(sigChan)
	case <-timer.C:
		sugar.Infow("Timer expired, stopping ingestor service...")
	}
}
