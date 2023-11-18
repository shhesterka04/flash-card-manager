package kafka

import (
	"errors"
	"fmt"
	"flash-card-manager/pkg/logger"
	"os"

	"github.com/joho/godotenv"
)

func LoadKafkaConfig() ([]string, error) {
	err := godotenv.Load("/home/art4m/prj/route256/flash-card-manager/.env")
	if err != nil {
		return nil, errors.New("error loading .env file")
	}

	kafkaHost := os.Getenv("KAFKA_HOST")
	if kafkaHost == "" {
		return nil, errors.New("KAFKA_HOST must be set in .env file")
	}

	kafkaPort := os.Getenv("KAFKA_PORT")
	if kafkaPort == "" {
		return nil, errors.New("KAFKA_PORT must be set in .env file")
	}

	return []string{fmt.Sprintf("%s:%s", kafkaHost, kafkaPort)}, nil
}

func InitializeKafka() (*Producer, *Consumer, error) {
	kafkaAddress, err := LoadKafkaConfig()
	if err != nil {
		logger.GetLogger().Sugar().Errorf("Error init kafka: %v", err)
	}

	producer, err := NewProducer(kafkaAddress)
	if err != nil {
		return nil, nil, err
	}

	consumer, err := NewConsumer(kafkaAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	go consumer.Consume("Card")
	go consumer.Consume("Deck")

	return producer, consumer, nil
}
