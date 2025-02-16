package ingestor

import (
	"encoding/json"
	"ingestor-service/internal/kafka"
	"ingestor-service/internal/websocket"
	"strconv"
	"time"

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

	err = s.client.Subscribe([]string{"BTC-USD"}, "market_trades")
	if err != nil {
		return err
	}

	messages := make(chan []byte, 1000000)

	go s.client.ReadMessages(messages, s.stop)

	go func() {

		for message := range messages {
			var messageData map[string]interface{}
			if err := json.Unmarshal(message, &messageData); err != nil {
				s.logger.Errorw("Failed to parse message", "error", err)
				continue
			}

			channel, _ := messageData["channel"].(string)
			if channel != "market_trades" {
				continue
			}

			events, ok := messageData["events"].([]interface{})
			if !ok {
				continue
			}

			for _, event := range events {
				eventMap, ok := event.(map[string]interface{})
				if !ok {
					continue
				}

				eventType, _ := eventMap["type"].(string)
				if eventType == "snapshot" || eventType == "update" {
					trades, ok := eventMap["trades"].([]interface{})
					if !ok {
						continue
					}

					for _, trade := range trades {
						tradeMap, ok := trade.(map[string]interface{})
						if !ok {
							continue
						}

						// Parse and reformat the timestamp
						timeStr, ok := tradeMap["time"].(string)
						if ok {
							if t, err := time.Parse(time.RFC3339Nano, timeStr); err == nil {
								// Format timestamp for Kafka/ksqlDB compatibility
								tradeMap["time"] = t.UnixMilli()
							}
						}

						// Convert price, size and trade_id to float
						priceStr, ok := tradeMap["price"].(string)
						if ok {
							price, err := strconv.ParseFloat(priceStr, 64)
							if err != nil {
								s.logger.Errorw("Failed to parse price", "error", err)
								continue
							}
							tradeMap["price"] = price
						}

						sizeStr, ok := tradeMap["size"].(string)
						if ok {
							size, err := strconv.ParseFloat(sizeStr, 64)
							if err != nil {
								s.logger.Errorw("Failed to parse size", "error", err)
								continue
							}
							tradeMap["size"] = size
						}

						tradeIDStr, ok := tradeMap["trade_id"].(string)
						if ok {
							tradeID, err := strconv.ParseFloat(tradeIDStr, 64)
							if err != nil {
								s.logger.Errorw("Failed to parse trade_id", "error", err)
								continue
							}
							tradeMap["trade_id"] = tradeID
						}

						// Create the simplified trade event
						tradeEvent, err := json.Marshal(tradeMap)
						if err != nil {
							s.logger.Errorw("Failed to marshal trade event", "error", err)
							continue
						}

						s.logger.Infow("Relaying trade event", "event", string(tradeEvent))
						err = s.RelayMessage(tradeEvent)
						if err != nil {
							s.logger.Errorw("Failed to relay message", "error", err)
						}
					}
				}
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
	s.logger.Infow("Stopping ingestor service...")
	s.stop <- true

	err := s.client.Disconnect()
	if err != nil {
		return err
	}
	s.logger.Infow("Ingestor service stopped")
	return nil
}
