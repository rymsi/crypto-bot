package kafka

import (
	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type Producer struct {
	syncProducer sarama.SyncProducer
	topic        string
	logger       *zap.SugaredLogger
}

func NewProducer(brokers []string, topic string, logger *zap.SugaredLogger) *Producer {
	syncProducer, err := sarama.NewSyncProducer(brokers, nil)
	if err != nil {
		logger.Errorw("Failed to create Kafka producer", "error", err)
		return nil
	}
	return &Producer{
		syncProducer: syncProducer,
		topic:        topic,
		logger:       logger,
	}

}

func (p *Producer) Produce(message []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.ByteEncoder(message),
	}

	partition, offset, err := p.syncProducer.SendMessage(msg)
	if err != nil {
		p.logger.Errorw("Failed to send message to Kafka", "error", err)
		return err
	}

	p.logger.Debugw("Message sent to Kafka", "partition", partition, "offset", offset)

	return nil

}

func (p *Producer) Close() error {
	return p.syncProducer.Close()
}
