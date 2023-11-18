//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"homework-3/internal/infrastructure/kafka"
	"homework-3/pkg/db"
	"homework-3/tests/postgres"
	"log"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type KafkaTestSuite struct {
	suite.Suite
	DB       *postgres.TDB
	Producer sarama.SyncProducer
	Consumer sarama.Consumer
	Topic    string
}

func (suite *KafkaTestSuite) SetupSuite() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	suite.DB = postgres.NewFromEnv(db.GenerateDsn())
	err = suite.DB.SetUp(suite.T())
	suite.Require().NoError(err)

	kafkaAddress, err := kafka.LoadKafkaConfig()
	if err != nil {
		log.Fatal(err)
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Consumer.Return.Errors = true

	suite.Producer, err = sarama.NewSyncProducer(kafkaAddress, config)
	suite.Require().NoError(err)

	suite.Consumer, err = sarama.NewConsumer(kafkaAddress, config)
	suite.Require().NoError(err)

	suite.Topic = "test-topic"
}

func (suite *KafkaTestSuite) TearDownSuite() {
	if err := suite.Producer.Close(); err != nil {
		log.Printf("Failed to close Kafka producer: %v", err)
	}

	if err := suite.Consumer.Close(); err != nil {
		log.Printf("Failed to close Kafka consumer: %v", err)
	}

	err := suite.DB.TearDown(context.Background(), suite.T())
	suite.Require().NoError(err)

	suite.CleanupKafka()
}

func (suite *KafkaTestSuite) CleanupKafka() {
	kafkaAddress, err := kafka.LoadKafkaConfig()
	if err != nil {
		log.Fatal(err)
	}

	admin, err := sarama.NewClusterAdmin(kafkaAddress, nil)
	if err != nil {
		log.Printf("Failed to create Kafka admin client: %v", err)
		return
	}
	defer func() {
		if err := admin.Close(); err != nil {
			log.Printf("Failed to close Kafka admin client: %v", err)
		}
	}()

	err = admin.DeleteTopic(suite.Topic)
	if err != nil {
		log.Printf("Failed to delete Kafka topic %s: %v", suite.Topic, err)
	} else {
		log.Printf("Kafka topic %s deleted successfully", suite.Topic)
	}
}

func (suite *KafkaTestSuite) TestKafkaIntegration() {
	topic := "test-topic"
	partition := int32(0)

	// Produce a message
	message := &sarama.ProducerMessage{
		Topic:     topic,
		Partition: partition,
		Value:     sarama.StringEncoder("Hello Kafka!"),
	}
	partition, offset, err := suite.Producer.SendMessage(message)
	suite.Require().NoError(err, "Failed to send message to Kafka")
	fmt.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", topic, partition, offset)

	// Consume the message
	partitionConsumer, err := suite.Consumer.ConsumePartition(topic, partition, sarama.OffsetOldest)
	suite.Require().NoError(err, "Failed to start consumer for partition %d", partition)
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Printf("Failed to close partition consumer: %v", err)
		}
	}()

	select {
	case msg := <-partitionConsumer.Messages():
		fmt.Printf("Consumed message offset %d\n", msg.Offset)
		suite.Equal("Hello Kafka!", string(msg.Value), "Unexpected message content")
	case err := <-partitionConsumer.Errors():
		suite.FailNow("Failed to consume message", err.Error())
	case <-time.After(5 * time.Second):
		suite.FailNow("Test timed out")
	}
}

func TestKafkaTestSuite(t *testing.T) {
	suite.Run(t, new(KafkaTestSuite))
}
