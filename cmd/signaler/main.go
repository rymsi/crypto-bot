package main

import (
	"context"
	"ingestor-service/internal/config"
	"ingestor-service/internal/signaler"
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

	cfg, err := config.LoadSignalerConfig(sugar)
	if err != nil {
		sugar.Errorw("Failed to load config", "error", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the Signaler Service
	signalerService, err := signaler.NewService(ctx, cfg.KafkaBrokers, cfg.KafkaConsumerGroup, sugar)
	if err != nil {
		sugar.Errorw("Failed to create signaler service", "error", err)
		return
	}

	// Start the Signaler Service
	err = signalerService.Start()
	if err != nil {
		sugar.Errorw("Failed to start signaler service", "error", err)
		return
	}
	defer signalerService.Stop()

	// Wait for SIGINT or SIGTERM or timer to run out
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	timer := time.NewTimer(60 * time.Second)

	select {
	case <-sigChan:
		sugar.Infow("Received SIGINT or SIGTERM, stopping ingestor service...")
		signal.Stop(sigChan)
	case <-timer.C:
		sugar.Infow("Timer expired, stopping ingestor service...")
		signal.Stop(sigChan)
	}
}
