//go:build integration
// +build integration

package tests

import (
	"context"
	"encoding/json"
	"homework-3/internal/infrastructure/kafka"
	"homework-3/pkg/db"
	"homework-3/pkg/repository/postgresql"
	"homework-3/pkg/repository/structs"
	"homework-3/tests/fixtures"
	"homework-3/tests/postgres"
	"log"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type DeckKafkaTestSuite struct {
	suite.Suite
	DB       *postgres.TDB
	Producer sarama.SyncProducer
	Consumer sarama.Consumer
	Topic    string
}

func (suite *DeckKafkaTestSuite) SetupSuite() {
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

	suite.Topic = "test-topic-decks"
}

func (suite *DeckKafkaTestSuite) TearDownSuite() {
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

func (suite *DeckKafkaTestSuite) CleanupKafka() {
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

func (suite *DeckKafkaTestSuite) TestKafkaPostDeck() {
	// Arrange
	ctx := context.Background()
	deckRepo := postgresql.NewDeck(suite.DB.DB)
	deck := fixtures.Deck().Valid().P()

	// Act
	deckID, err := deckRepo.Add(ctx, *deck)
	suite.Require().NoError(err)

	deck.ID = deckID
	deckBytes, err := json.Marshal(deck)
	suite.Require().NoError(err)

	message := &sarama.ProducerMessage{
		Topic: suite.Topic,
		Value: sarama.ByteEncoder(deckBytes),
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
		suite.Equal(deckBytes, msg.Value, "Unexpected message content")
	case err := <-partitionConsumer.Errors():
		suite.FailNow("Failed to consume message", err.Error())
	case <-time.After(5 * time.Second):
		suite.FailNow("Test timed out")
	}
}

func (suite *DeckKafkaTestSuite) TestKafkaUpdateDeck() {
	// Arrange
	ctx := context.Background()
	deckRepo := postgresql.NewDeck(suite.DB.DB)
	deck := fixtures.Deck().Valid().P()

	// Act
	deckID, err := deckRepo.Add(ctx, *deck)
	deckID, err = deckRepo.Update(ctx, *deck)
	suite.Require().NoError(err)

	deck.ID = deckID
	deckBytes, err := json.Marshal(deck)
	suite.Require().NoError(err)

	message := &sarama.ProducerMessage{
		Topic: suite.Topic,
		Value: sarama.ByteEncoder(deckBytes),
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
		suite.Equal(deckBytes, msg.Value, "Unexpected message content")
	case err := <-partitionConsumer.Errors():
		suite.FailNow("Failed to consume message", err.Error())
	case <-time.After(5 * time.Second):
		suite.FailNow("Test timed out")
	}
}

func (suite *DeckKafkaTestSuite) TestKafkaDeleteDeck() {
	// Arrange
	ctx := context.Background()
	deckRepo := postgresql.NewDeck(suite.DB.DB)
	deck := fixtures.Deck().Valid().P()

	// Act
	deckID, err := deckRepo.Add(ctx, *deck)
	err = deckRepo.Delete(ctx, deckID)
	suite.Require().NoError(err)

	deck.ID = deckID
	deckBytes, err := json.Marshal(deck)
	suite.Require().NoError(err)

	message := &sarama.ProducerMessage{
		Topic: suite.Topic,
		Value: sarama.ByteEncoder(deckBytes),
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
		suite.Equal(deckBytes, msg.Value, "Unexpected message content")
	case err := <-partitionConsumer.Errors():
		suite.FailNow("Failed to consume message", err.Error())
	case <-time.After(5 * time.Second):
		suite.FailNow("Test timed out")
	}
}

func (suite *DeckKafkaTestSuite) TestKafkaGetDeck() {
	// Arrange
	ctx := context.Background()
	deckRepo := postgresql.NewDeck(suite.DB.DB)
	expectedDeck := fixtures.Deck().Valid().P()

	// Act
	deckID, err := deckRepo.Add(ctx, *expectedDeck)
	suite.Require().NoError(err)

	expectedDeck.ID = deckID
	expectedDeckBytes, err := json.Marshal(expectedDeck)
	suite.Require().NoError(err)

	message := &sarama.ProducerMessage{
		Topic: suite.Topic,
		Value: sarama.ByteEncoder(expectedDeckBytes),
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
		suite.Equal(expectedDeckBytes, msg.Value, "Unexpected message content")
	case err := <-partitionConsumer.Errors():
		suite.FailNow("Failed to consume message", err.Error())
	case <-time.After(5 * time.Second):
		suite.FailNow("Test timed out")
	}

	retrievedDeck, err := deckRepo.GetByID(ctx, deckID)
	suite.Require().NoError(err)
	opt := cmpopts.IgnoreFields(structs.Deck{}, "CreatedAt")
	if !cmp.Equal(expectedDeck, retrievedDeck, opt) {
		suite.FailNow("Retrieved deck does not match expected deck", cmp.Diff(expectedDeck, retrievedDeck, opt))
	}
}

func TestDeckKafkaTestSuite(t *testing.T) {
	suite.Run(t, new(DeckKafkaTestSuite))
}
