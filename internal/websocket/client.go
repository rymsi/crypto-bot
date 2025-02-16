package websocket

import (
	"encoding/json"
	"strconv"
	"time"

	"ingestor-service/internal/config"

	"github.com/gorilla/websocket"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

type Client struct {
	url    string
	conn   *websocket.Conn
	logger *zap.SugaredLogger
	config *config.IngestorConfig
}

func NewClient(url string, config *config.IngestorConfig, logger *zap.SugaredLogger) *Client {
	return &Client{
		url:    url,
		logger: logger,
		config: config,
	}
}

func (c *Client) Connect() error {
	c.logger.Infow("Connecting to Coinbase WebSocket...")
	conn, response, err := websocket.DefaultDialer.Dial(c.url, nil)

	if err != nil {
		c.logger.Errorw("Failed to connect to Coinbase WebSocket", "error", err)
		return err
	}

	c.logger.Infow("Connected to Coinbase WebSocket", "response", response.Body)
	c.conn = conn
	return nil

}

func (c *Client) Subscribe(productIds []string, channel string) error {

	message := map[string]interface{}{
		"type":        "subscribe",
		"product_ids": productIds,
		"channel":     channel,
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	c.logger.Infow("Sending subscribe message", "message", string(messageBytes))
	err = c.conn.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		c.logger.Errorw("Failed to subscribe to Coinbase WebSocket", "error", err)
		return err
	}

	return nil
}

func (c *Client) ReadMessages(messages chan<- []byte, stopchan <-chan bool) error {
	defer close(messages)

	sequenceCache := cache.New(1*time.Minute, 2*time.Minute)

	for {
		select {
		case <-stopchan:
			c.logger.Infow("Stopping read messages from Coinbase WebSocket...")
			return nil
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				c.logger.Errorw("Failed to read message from Coinbase WebSocket", "error", err)
				return err
			}

			// Parse message to get sequence
			var data map[string]interface{}
			if err := json.Unmarshal(message, &data); err != nil {
				c.logger.Errorw("Failed to parse message", "error", err)
				continue
			}

			// Check if sequence exists in message
			if seq, ok := data["sequence"].(float64); ok {
				// Convert to int64 for cache key
				seqInt := int64(seq)

				// Check if sequence already seen
				if _, found := sequenceCache.Get(strconv.FormatInt(seqInt, 10)); found {
					c.logger.Infow("Skipping duplicate sequence", "sequence", seqInt)
					continue
				}

				// Store sequence in cache
				sequenceCache.Set(strconv.FormatInt(seqInt, 10), true, 1*time.Minute)
			}

			c.logger.Infow("Received message from Coinbase WebSocket", "message", string(message))
			messages <- message
		}
	}
}

func (c *Client) Disconnect() error {

	c.logger.Infow("Disconnecting from Coinbase WebSocket...")

	_ = c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	err := c.conn.Close()
	if err != nil {
		c.logger.Errorw("Failed to disconnect from Coinbase WebSocket", "error", err)
		return err
	}

	c.logger.Infow("Disconnected from Coinbase WebSocket")
	return nil
}
