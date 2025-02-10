package signaler

import (
	"context"
	"encoding/json"
	"ingestor-service/internal/kafka"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type Service struct {
	ctx          context.Context
	producer     *kafka.Producer
	eventChannel chan []byte
	consumer     *kafka.Consumer
	logger       *zap.SugaredLogger
}

func NewService(ctx context.Context, brokers []string, groupID string, logger *zap.SugaredLogger) (*Service, error) {
	service := &Service{
		ctx:          ctx,
		logger:       logger,
		eventChannel: make(chan []byte, 1000),
	}

	consumerKafkaTopic := "btc-usd"
	producerKafkaTopic := "btc-usd-signals"

	consumer, err := kafka.NewConsumer(ctx, brokers, consumerKafkaTopic, groupID, nil, logger)
	if err != nil {
		return nil, err
	}

	producer, err := kafka.NewProducer(brokers, producerKafkaTopic, logger)
	if err != nil {
		logger.Errorw("Failed to create kafka producer", "error", err)
		return nil, err
	}

	service.producer = producer
	service.consumer = consumer

	consumer.SetMessageHandler(service.handler)

	return service, nil
}

func (s *Service) Start() error {
	s.logger.Infow("Starting signaler service...")

	err := s.consumer.Consume()
	if err != nil {
		s.logger.Errorw("Failed to consume messages", "error", err)
		return err
	}

	<-s.consumer.Ready()

	s.produceSignals()

	return nil
}

func (s *Service) Stop() {
	s.logger.Infow("Stopping signaler service...")
	s.consumer.Close()
	s.producer.Close()
	close(s.eventChannel)
}

func (s *Service) handler(ctx context.Context, msg []byte) error {
	s.logger.Debugw("Signaler service handler received message")

	s.eventChannel <- msg

	return nil
}

func (s *Service) produceSignals() {
	go func() {
		buffer := make([]map[string]interface{}, 0, 10)
		retryCount := 0
		backoffTime := time.Duration(5) * time.Millisecond
		maxRetries := 100

		for {
			select {
			case msg, ok := <-s.eventChannel:
				if !ok {
					s.logger.Infow("Event channel closed, exiting...")
					return
				}

				// s.logger.Infow("Signaler service producer received message")

				// Reset retry count on successful read
				retryCount = 0
				backoffTime = time.Duration(5) * time.Millisecond

				eventRaw := make(map[string]interface{})
				json.Unmarshal(msg, &eventRaw)

				event := make(map[string]interface{})
				event["price"] = eventRaw["price"]
				event["timestamp"] = eventRaw["time"]

				buffer = append(buffer, event)

				if len(buffer) == 10 {
					avgPrice := 0.0
					for _, e := range buffer {
						priceStr, _ := e["price"].(string)
						price, err := strconv.ParseFloat(priceStr, 64)
						if err != nil {
							s.logger.Errorw("Failed to parse price", "error", err)
							continue
						}
						avgPrice += price
					}
					avgPrice = avgPrice / float64(len(buffer))

					signal := map[string]interface{}{
						"avgPrice":  avgPrice,
						"timestamp": buffer[4]["timestamp"],
					}

					signalJSON, err := json.Marshal(signal)
					if err != nil {
						s.logger.Errorw("Failed to marshal signal", "error", err)
					} else {
						s.logger.Infow("Buffer full so sending buffer to Produce()")
						go func() {
							if err := s.producer.Produce(signalJSON); err != nil {
								s.logger.Errorw("Failed to produce signal", "error", err)
							}
						}()
					}

					// Reset buffer
					buffer = buffer[:0]
				} else {
					s.logger.Infow("Buffer not full", "buffer_size", len(buffer))
				}

			default:
				if retryCount < maxRetries {
					retryCount++
					s.logger.Infow("Event channel empty", "attempt", retryCount, "backoffTime", backoffTime)
					backoffTime *= 2
					time.Sleep(backoffTime)

				} else {
					s.logger.Errorw("Max retries reached with empty event channel")
					return
				}
			}
		}
	}()
}
