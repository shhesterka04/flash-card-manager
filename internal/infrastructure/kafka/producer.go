//go:generate mockgen -source ./producer.go -destination=./mocks/kafka.go -package=mock_kafka
package kafka

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/pkg/errors"
)

type Producer struct {
	brokers      []string
	syncProducer sarama.SyncProducer
}

type ProducerInterface interface {
	SendSyncMessage(message *sarama.ProducerMessage) (partition int32, offset int64, err error)
	SendSyncMessages(messages []*sarama.ProducerMessage) error
	Close() error
}

func NewProducer(brokers []string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	syncProducer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, errors.Wrap(err, "error with sync kafka-producer")
	}

	producer := &Producer{
		brokers:      brokers,
		syncProducer: syncProducer,
	}

	return producer, nil
}

func (k *Producer) SendSyncMessage(message *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	return k.syncProducer.SendMessage(message)
}

func (k *Producer) SendSyncMessages(messages []*sarama.ProducerMessage) error {
	err := k.syncProducer.SendMessages(messages)
	if err != nil {
		fmt.Println("kafka.Producer.SendMessages error", err)
	}

	return err
}

func (k *Producer) Close() error {
	err := k.syncProducer.Close()
	if err != nil {
		return errors.Wrap(err, "kafka.Producer.Close syncProducer")
	}

	return nil
}
