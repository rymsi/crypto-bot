package main

import (
	"context"
	"ingestor-service/internal/bot"
	"ingestor-service/internal/config"
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

	cfg, err := config.LoadBotConfig(sugar)
	if err != nil {
		sugar.Errorw("Failed to load config", "error", err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the Bot Service
	botService, err := bot.NewService(ctx, cfg.KafkaBrokers, cfg.KafkaConsumerGroup, sugar)
	if err != nil {
		sugar.Errorw("Failed to create bot service", "error", err)
		return
	}

	// Start the Bot Service
	err = botService.Start()
	if err != nil {
		sugar.Errorw("Failed to start bot service", "error", err)
		return
	}

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
	}

	botService.Stop()
}
