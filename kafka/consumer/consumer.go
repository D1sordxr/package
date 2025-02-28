package consumer

import "github.com/segmentio/kafka-go"

type Config struct {
	Brokers []string
}

// TODO: NewConsumer

func NewConsumer(config Config) (r *kafka.Reader, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	// TODO: ...

	//reader := kafka.NewReader(kafka.ReaderConfig{
	//	Brokers: config.Brokers,
	//	GroupID: "group-id",
	//	Topic:   "topic",
	//})

	//return reader, err
	return nil, nil
}
