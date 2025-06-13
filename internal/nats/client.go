package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"test-task-one/internal/models/ch"
)

type NATSClient struct {
	conn *nats.Conn
}

func NewNATSClient(url string) (*NATSClient, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}
	return &NATSClient{conn: nc}, nil
}

func (c *NATSClient) PublishEvent(ctx context.Context, event *ch.Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = c.conn.Publish("goods.events", payload)
	if err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}

func (c *NATSClient) Close() {
	c.conn.Close()
}
