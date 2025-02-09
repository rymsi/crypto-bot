package ingestor

import (
	"ingestor-service/internal/kafka"
	"ingestor-service/internal/websocket"

	"go.uber.org/zap"
)

type Service struct {
	logger   *zap.SugaredLogger
	client   *websocket.Client
	stop     chan bool
	producer *kafka.Producer
}

func NewService(client *websocket.Client, producer *kafka.Producer, logger *zap.SugaredLogger) *Service {

	return &Service{
		client:   client,
		producer: producer,
		stop:     make(chan bool),
		logger:   logger,
	}

}

func (s *Service) Start() error {
	err := s.client.Connect()
	if err != nil {
		return err
	}

	err = s.client.Subscribe([]string{"BTC-USD"}, []string{"ticker"})
	if err != nil {
		return err
	}

	messages := make(chan []byte, 100)

	go s.client.ReadMessages(messages, s.stop)

	go func() {
		// Discard the first response message which is the subscription response
		<-messages

		for message := range messages {
			s.logger.Infow("Relaying message: ", "info", string(message))
			err := s.RelayMessage(message)
			if err != nil {
				s.logger.Errorw("Failed to relay message", "error", err)
			}
		}
	}()

	return nil
}

func (s *Service) RelayMessage(message []byte) error {
	err := s.producer.Produce(message)
	if err != nil {
		s.logger.Errorw("Failed to produce message to Kafka", "error", err)
		return err
	}
	return nil
}

func (s *Service) Stop() error {
	s.stop <- true

	err := s.client.Disconnect()
	if err != nil {
		return err
	}

	return nil
}
