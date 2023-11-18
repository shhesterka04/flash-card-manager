//go:build integration
// +build integration

package tests

import (
	"context"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
	"homework-3/pkg/db"
	"homework-3/pkg/repository/postgresql"
	"homework-3/tests/fixtures"
	"homework-3/tests/postgres"
	"log"
	"testing"
)

type CardTestSuite struct {
	suite.Suite
	DB *postgres.TDB
}

func (suite *CardTestSuite) SetupTest() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	suite.DB = postgres.NewFromEnv(db.GenerateDsn())
	suite.Require().NoError(err)

	err = suite.DB.SetUp(suite.T())
	suite.Require().NoError(err)
}

func (suite *CardTestSuite) TearDownTest() {
	ctx := context.Background()
	err := suite.DB.TearDown(ctx, suite.T())
	suite.Require().NoError(err)
}

func (suite *CardTestSuite) TestCreateCard() {
	ctx := context.Background()
	cardRepo := postgresql.NewCard(suite.DB.DB)
	deckRepo := postgresql.NewDeck(suite.DB.DB)

	deckValid := fixtures.Deck().Valid().P()
	deckID, err := deckRepo.Add(ctx, *deckValid)
	suite.Require().NoError(err)

	cardValid := fixtures.Card().Valid().DeckID(deckID).P()

	// Act
	cardID, err := cardRepo.Add(ctx, *cardValid)
	suite.Require().NoError(err)

	cardFromDB, err := cardRepo.GetByID(ctx, cardID)
	suite.Require().NoError(err)

	suite.Assert().Equal(cardValid.Front, cardFromDB.Front)
	suite.Assert().Equal(cardValid.Back, cardFromDB.Back)
	suite.Assert().Equal(deckID, cardFromDB.DeckID)
}

func (suite *CardTestSuite) TestDeleteCard() {
	// Arrange
	ctx := context.Background()
	cardRepo := postgresql.NewCard(suite.DB.DB)
	deckRepo := postgresql.NewDeck(suite.DB.DB)

	deckValid := fixtures.Deck().Valid().P()
	deckID, err := deckRepo.Add(ctx, *deckValid)
	suite.Require().NoError(err)

	cardValid := fixtures.Card().Valid().DeckID(deckID).P()
	cardID, err := cardRepo.Add(ctx, *cardValid)
	suite.Require().NoError(err)

	// Act
	err = cardRepo.Delete(ctx, cardID)

	// Assert
	suite.Require().NoError(err)
	_, err = cardRepo.GetByID(ctx, cardID)
	suite.Assert().Error(err)
}

func (suite *CardTestSuite) TestGetByIDCard() {
	// Arrange
	ctx := context.Background()
	cardRepo := postgresql.NewCard(suite.DB.DB)
	deckRepo := postgresql.NewDeck(suite.DB.DB)

	deckValid := fixtures.Deck().Valid().P()
	deckID, err := deckRepo.Add(ctx, *deckValid)
	suite.Require().NoError(err)

	cardValid := fixtures.Card().Valid().DeckID(deckID).P()
	cardID, err := cardRepo.Add(ctx, *cardValid)
	suite.Require().NoError(err)

	// Act
	cardFromDB, err := cardRepo.GetByID(ctx, cardID)

	// Assert
	suite.Require().NoError(err)
	suite.Assert().Equal(cardValid.Front, cardFromDB.Front)
	suite.Assert().Equal(cardValid.Back, cardFromDB.Back)
	suite.Assert().Equal(deckID, cardFromDB.DeckID)
}

func (suite *CardTestSuite) TestUpdateCard() {
	// Arrange
	ctx := context.Background()
	cardRepo := postgresql.NewCard(suite.DB.DB)
	deckRepo := postgresql.NewDeck(suite.DB.DB)

	deckValid := fixtures.Deck().Valid().P()
	deckID, err := deckRepo.Add(ctx, *deckValid)
	suite.Require().NoError(err)

	cardValid := fixtures.Card().Valid().DeckID(deckID).P()
	cardID, err := cardRepo.Add(ctx, *cardValid)
	suite.Require().NoError(err)

	updatedCard := *cardValid
	updatedCard.ID = cardID
	updatedCard.Front = "Updated Front"
	updatedCard.Back = "Updated Back"

	// Act
	_, err = cardRepo.Update(ctx, updatedCard)

	// Assert
	suite.Require().NoError(err)
	cardFromDB, err := cardRepo.GetByID(ctx, cardID)
	suite.Require().NoError(err)
	suite.Assert().Equal(updatedCard.Front, cardFromDB.Front)
	suite.Assert().Equal(updatedCard.Back, cardFromDB.Back)
}

func TestCardTestSuite(t *testing.T) {
	suite.Run(t, new(CardTestSuite))
}
