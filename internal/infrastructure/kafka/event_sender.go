package kafka

import (
	"encoding/json"

	"github.com/IBM/sarama"
)

type EventSender interface {
	SendEvent(eventType, query string) error
}

type KafkaEventSender struct {
	producer ProducerInterface
}

func NewKafkaEventSender(producer ProducerInterface) *KafkaEventSender {
	return &KafkaEventSender{producer: producer}
}

func (s *KafkaEventSender) SendEvent(eventType, query string) error {

	event := NewEvent(eventType, query)
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return err
	}
	msg := &sarama.ProducerMessage{
		Topic: "Card",
		Value: sarama.ByteEncoder(eventBytes),
	}

	_, _, err = s.producer.SendSyncMessage(msg)
	if err != nil {
		return err
	}
	return nil
}
