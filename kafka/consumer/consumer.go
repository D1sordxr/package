package consumer

import (
	"github.com/segmentio/kafka-go"
)

type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

func NewConsumer(config *Config) *kafka.Reader {
	if len(config.Brokers) == 0 || config.Topic == "" || config.GroupID == "" {
		panic("invalid config data")
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: config.Brokers,
		GroupID: config.GroupID,
		Topic:   config.Topic,
	})

	return reader
}
