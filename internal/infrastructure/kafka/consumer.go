package kafka

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

type Consumer struct {
	brokers        []string
	SingleConsumer sarama.Consumer
}

func NewConsumer(brokers []string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = false
	config.Consumer.Offsets.AutoCommit.Enable = true
	config.Consumer.Offsets.AutoCommit.Interval = 5 * time.Second

	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumer(brokers, config)

	if err != nil {
		return nil, err
	}

	return &Consumer{
		brokers:        brokers,
		SingleConsumer: consumer,
	}, err

}

func (c *Consumer) Consume(topic string) {
	partitionConsumer, err := c.SingleConsumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		fmt.Println("Error starting the partition consumer:", err)
		return
	}
	defer partitionConsumer.Close()

	for {
		select {
		case msg := <-partitionConsumer.Messages():
			c.handleMessage(msg)
		case err := <-partitionConsumer.Errors():
			fmt.Println("Error while consuming from Kafka:", err)
		}
	}
}

func (c *Consumer) handleMessage(msg *sarama.ConsumerMessage) {
	var event Event
	err := json.Unmarshal(msg.Value, &event)
	if err != nil {
		fmt.Println("Error unmarshalling Kafka message:", err)
		return
	}
	fmt.Printf("Received message at %v of type %s: %s\n", event.Timestamp, event.Type, event.Query)
}
