//go:build integration
// +build integration

package tests

import (
	"context"
	"encoding/json"
	"homework-3/internal/infrastructure/kafka"
	"homework-3/pkg/db"
	"homework-3/pkg/repository/postgresql"
	"homework-3/tests/fixtures"
	"homework-3/tests/postgres"
	"log"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type CardKafkaTestSuite struct {
	suite.Suite
	DB       *postgres.TDB
	Producer sarama.SyncProducer
	Consumer sarama.Consumer
	Topic    string
}

func (suite *CardKafkaTestSuite) SetupSuite() {
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

	suite.Topic = "test-topic-card"
}

func (suite *CardKafkaTestSuite) TearDownSuite() {
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

func (suite *CardKafkaTestSuite) CleanupKafka() {
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

func (suite *CardKafkaTestSuite) TestKafkaPost() {
	// Arrange
	ctx := context.Background()
	cardRepo := postgresql.NewCard(suite.DB.DB)
	deckRepo := postgresql.NewDeck(suite.DB.DB)
	deck := fixtures.Deck().Valid().P()
	deckID, err := deckRepo.Add(ctx, *deck)
	suite.Require().NoError(err)
	card := fixtures.Card().Valid().DeckID(deckID).P()

	// Act
	cardID, err := cardRepo.Add(ctx, *card)
	suite.Require().NoError(err)

	card.ID = cardID
	cardBytes, err := json.Marshal(card)
	suite.Require().NoError(err)

	message := &sarama.ProducerMessage{
		Topic: suite.Topic,
		Value: sarama.ByteEncoder(cardBytes),
	}

	partition, offset, err := suite.Producer.SendMessage(message)
	suite.Require().NoError(err)
	log.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", suite.Topic, partition, offset)

	partitionConsumer, err := suite.Consumer.ConsumePartition(suite.Topic, partition, offset)
	suite.Require().NoError(err)
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Printf("Failed to close partition consumer: %v", err)
		}
	}()

	select {
	case msg := <-partitionConsumer.Messages():
		suite.Equal(cardBytes, msg.Value, "Unexpected message content")
	case err := <-partitionConsumer.Errors():
		suite.FailNow("Failed to consume message", err.Error())
	case <-time.After(5 * time.Second):
		suite.FailNow("Test timed out")
	}
}

func (suite *CardKafkaTestSuite) TestKafkaPut() {
	// Arrange
	ctx := context.Background()
	cardRepo := postgresql.NewCard(suite.DB.DB)
	card := fixtures.Card().Valid().DeckID(1).P()

	// Act
	cardID, err := cardRepo.Update(ctx, *card)
	suite.Require().NoError(err)

	card.ID = cardID
	cardBytes, err := json.Marshal(card)
	suite.Require().NoError(err)

	message := &sarama.ProducerMessage{
		Topic: suite.Topic,
		Value: sarama.ByteEncoder(cardBytes),
	}

	partition, offset, err := suite.Producer.SendMessage(message)
	suite.Require().NoError(err)
	log.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", suite.Topic, partition, offset)

	partitionConsumer, err := suite.Consumer.ConsumePartition(suite.Topic, partition, offset)
	suite.Require().NoError(err)
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Printf("Failed to close partition consumer: %v", err)
		}
	}()

	select {
	case msg := <-partitionConsumer.Messages():
		suite.Equal(cardBytes, msg.Value, "Unexpected message content")
	case err := <-partitionConsumer.Errors():
		suite.FailNow("Failed to consume message", err.Error())
	case <-time.After(5 * time.Second):
		suite.FailNow("Test timed out")
	}
}

func (suite *CardKafkaTestSuite) TestKafkaDelete() {
	// Arrange
	ctx := context.Background()
	cardRepo := postgresql.NewCard(suite.DB.DB)
	card := fixtures.Card().Valid().DeckID(1).P()

	// Act
	err := cardRepo.Delete(ctx, 1)
	suite.Require().NoError(err)

	cardBytes, err := json.Marshal(card)
	suite.Require().NoError(err)

	message := &sarama.ProducerMessage{
		Topic: suite.Topic,
		Value: sarama.ByteEncoder(cardBytes),
	}

	partition, offset, err := suite.Producer.SendMessage(message)
	suite.Require().NoError(err)
	log.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", suite.Topic, partition, offset)

	partitionConsumer, err := suite.Consumer.ConsumePartition(suite.Topic, partition, offset)
	suite.Require().NoError(err)
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Printf("Failed to close partition consumer: %v", err)
		}
	}()

	select {
	case msg := <-partitionConsumer.Messages():
		suite.Equal(cardBytes, msg.Value, "Unexpected message content")
	case err := <-partitionConsumer.Errors():
		suite.FailNow("Failed to consume message", err.Error())
	case <-time.After(5 * time.Second):
		suite.FailNow("Test timed out")
	}
}

func (suite *CardKafkaTestSuite) TestKafkaGet() {
	// Arrange
	ctx := context.Background()
	cardRepo := postgresql.NewCard(suite.DB.DB)
	card := fixtures.Card().Valid().DeckID(1).P()

	// Act
	err := cardRepo.Delete(ctx, 1)
	suite.Require().NoError(err)

	cardBytes, err := json.Marshal(card)
	suite.Require().NoError(err)

	message := &sarama.ProducerMessage{
		Topic: suite.Topic,
		Value: sarama.ByteEncoder(cardBytes),
	}

	partition, offset, err := suite.Producer.SendMessage(message)
	suite.Require().NoError(err)
	log.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", suite.Topic, partition, offset)

	partitionConsumer, err := suite.Consumer.ConsumePartition(suite.Topic, partition, offset)
	suite.Require().NoError(err)
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Printf("Failed to close partition consumer: %v", err)
		}
	}()

	select {
	case msg := <-partitionConsumer.Messages():
		suite.Equal(cardBytes, msg.Value, "Unexpected message content")
	case err := <-partitionConsumer.Errors():
		suite.FailNow("Failed to consume message", err.Error())
	case <-time.After(5 * time.Second):
		suite.FailNow("Test timed out")
	}
}

func TestCardKafkaTestSuite(t *testing.T) {
	suite.Run(t, new(CardKafkaTestSuite))
}
