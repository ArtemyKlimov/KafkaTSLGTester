package kafka

import (
	"fmt"
	"log/slog"

	"github.com/IBM/sarama"
)

// Producer wraps sarama.AsyncProducer with a background error drain goroutine.
type Producer struct {
	ap   sarama.AsyncProducer
	done chan struct{}
}

// NewProducer creates an AsyncProducer and starts the error drain loop.
func NewProducer(brokers []string, cfg *sarama.Config) (*Producer, error) {
	ap, err := sarama.NewAsyncProducer(brokers, cfg)
	if err != nil {
		return nil, fmt.Errorf("creating async producer: %w", err)
	}
	p := &Producer{ap: ap, done: make(chan struct{})}
	go p.drainErrors()
	return p, nil
}

// Send enqueues a message for async delivery. Thread-safe.
// Если key пустой — сообщение отправляется без ключа (nil), Kafka распределяет по round-robin.
func (p *Producer) Send(topic, key string, value []byte) {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(value),
	}
	if key != "" {
		msg.Key = sarama.StringEncoder(key)
	}
	p.ap.Input() <- msg
}

// Close flushes remaining messages and waits for the drain loop to exit.
func (p *Producer) Close() error {
	err := p.ap.Close()
	<-p.done
	return err
}

func (p *Producer) drainErrors() {
	defer close(p.done)
	for err := range p.ap.Errors() {
		slog.Error("kafka send error", "topic", err.Msg.Topic, "err", err.Err)
	}
}
