package websocket

import (
	"errors"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Client struct {
	url    string
	conn   *websocket.Conn
	logger *zap.SugaredLogger
}

func NewClient(url string, logger *zap.SugaredLogger) *Client {
	return &Client{
		url:    url,
		logger: logger,
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

func (c *Client) Subscribe(productIds []string, channels []string) error {
	if c.conn == nil {
		return errors.New("not connected to Coinbase WebSocket")
	}

	message := map[string]interface{}{
		"type":        "subscribe",
		"product_ids": productIds,
		"channels":    channels,
	}

	err := c.conn.WriteJSON(message)
	if err != nil {
		c.logger.Errorw("Failed to subscribe to Coinbase WebSocket", "error", err)
		return err
	}
	c.logger.Infow("Subscribed to Coinbase WebSocket")
	return nil
}

func (c *Client) ReadMessages(messages chan<- []byte, stopchan <-chan bool) error {
	defer close(messages)

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
