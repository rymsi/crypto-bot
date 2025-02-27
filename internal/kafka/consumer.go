package kafka

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type MessageHandler func(ctx context.Context, msg []byte) error

type Consumer struct {
	ctx           context.Context
	consumerGroup sarama.ConsumerGroup
	topic         string
	cgHandler     *consumerGroupHandler
	ready         chan bool
	logger        *zap.SugaredLogger
}

func NewConsumer(ctx context.Context, brokers []string, topic string, group string, handler MessageHandler, logger *zap.SugaredLogger) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()

	consumerGroup, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		logger.Errorw("Failed to create consumer group", "error", err)
		return nil, err
	}

	ready := make(chan bool)

	cgHandler := &consumerGroupHandler{
		handler: handler,
		logger:  logger,
		ready:   ready,
	}

	return &Consumer{
		ctx:           ctx,
		consumerGroup: consumerGroup,
		topic:         topic,
		cgHandler:     cgHandler,
		ready:         ready,
		logger:        logger,
	}, nil
}

func (c *Consumer) Consume() error {
	go func() {
		defer func() {
			err := c.consumerGroup.Close()
			if err != nil {
				c.logger.Errorw("Failed to close consumer group", "error", err)
			}
		}()

		err := c.consumerGroup.Consume(c.ctx, []string{c.topic}, c.cgHandler)
		if err != nil {
			c.logger.Errorw("Failed to consume messages", "error", err)
		}

		if c.ctx.Err() != nil {
			c.logger.Errorw("Consumer context error", "error", c.ctx.Err())
		}
	}()

	return nil

}

func (c *Consumer) Close() error {
	return c.consumerGroup.Close()
}

func (c *Consumer) SetMessageHandler(handler MessageHandler) {
	c.cgHandler.handler = handler
}

func (c *Consumer) Ready() <-chan bool {
	return c.ready
}

type consumerGroupHandler struct {
	handler MessageHandler
	logger  *zap.SugaredLogger
	ready   chan bool
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	h.logger.Info("Consumer group handler setup")
	close(h.ready)
	return nil
}

func (h *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	h.logger.Info("Consumer group handler cleanup")
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	for message := range claim.Messages() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := h.handler(ctx, message.Value)
		// h.logger.Infow("Message consumed")
		if err != nil {
			h.logger.Errorf("Error handling message: %s", err)
		} else {
			session.MarkMessage(message, "")
		}

		cancel()
	}
	return nil
}
