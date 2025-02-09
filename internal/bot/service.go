package bot

import (
	"context"
	"ingestor-service/internal/kafka"

	"go.uber.org/zap"
)

type Service struct {
	ctx      context.Context
	consumer *kafka.Consumer
	logger   *zap.SugaredLogger
}

func NewService(ctx context.Context, brokers []string, consumerGroup string, logger *zap.SugaredLogger) (*Service, error) {
	service := &Service{ctx: ctx, logger: logger}

	kafkaTopic := "btc-usd"
	consumer, err := kafka.NewConsumer(ctx, brokers, kafkaTopic, consumerGroup, service.handleMessage, logger)
	if err != nil {
		return nil, err
	}
	service.consumer = consumer

	return service, nil
}

func (s *Service) Start() error {
	err := s.consumer.Consume()
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) Stop() error {
	s.consumer.Close()
	return nil
}

func (s *Service) handleMessage(ctx context.Context, msg []byte) error {
	s.logger.Infof("Received message: %s", string(msg))
	return nil
}
