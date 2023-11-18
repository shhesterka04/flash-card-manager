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

type DeckTestSuite struct {
	suite.Suite
	DB *postgres.TDB
}

func (suite *DeckTestSuite) SetupTest() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	suite.DB = postgres.NewFromEnv(db.GenerateDsn())
	suite.Require().NoError(err)

	err = suite.DB.SetUp(suite.T())
	suite.Require().NoError(err)

}

func (suite *DeckTestSuite) TearDownTest() {
	ctx := context.Background()
	suite.DB.TearDown(ctx, suite.T())
}

func (suite *DeckTestSuite) TestAddDeck() {
	// Arrange
	ctx := context.Background()
	deckRepo := postgresql.NewDeck(suite.DB.DB)
	deckValid := fixtures.Deck().Valid().P()

	// Act
	deckID, err := deckRepo.Add(ctx, *deckValid)

	// Assert
	suite.Require().NoError(err)
	suite.Assert().NotZero(deckID)
}

func (suite *DeckTestSuite) TestDeleteDeck() {
	// Arrange
	ctx := context.Background()
	deckRepo := postgresql.NewDeck(suite.DB.DB)
	deckValid := fixtures.Deck().Valid().P()
	deckID, err := deckRepo.Add(ctx, *deckValid)
	suite.Require().NoError(err)

	// Act
	err = deckRepo.Delete(ctx, deckID)

	// Assert
	suite.Require().NoError(err)

	_, err = deckRepo.GetByID(ctx, deckID)
	suite.Assert().Error(err)
}

func (suite *DeckTestSuite) TestGetByIDDeck() {
	// Arrange
	ctx := context.Background()
	deckRepo := postgresql.NewDeck(suite.DB.DB)
	deckValid := fixtures.Deck().Valid().P()
	deckID, err := deckRepo.Add(ctx, *deckValid)
	suite.Require().NoError(err)

	// Act
	deckFromDB, err := deckRepo.GetByID(ctx, deckID)

	// Assert
	suite.Require().NoError(err)
	suite.Assert().Equal(deckValid.Title, deckFromDB.Title)
	suite.Assert().Equal(deckValid.Description, deckFromDB.Description)
	suite.Assert().Equal(deckValid.Author, deckFromDB.Author)
}

func (suite *DeckTestSuite) TestUpdateDeck() {
	// Arrange
	ctx := context.Background()
	deckRepo := postgresql.NewDeck(suite.DB.DB)
	deckValid := fixtures.Deck().Valid().P()
	deckID, err := deckRepo.Add(ctx, *deckValid)
	suite.Require().NoError(err)

	updatedDeck := *deckValid
	updatedDeck.ID = deckID
	updatedDeck.Title = "Updated Title"
	updatedDeck.Description = "Updated Description"
	updatedDeck.Author = "Updated Author"

	// Act
	affectedRows, err := deckRepo.Update(ctx, updatedDeck)

	// Assert
	suite.Require().NoError(err)
	suite.Assert().Equal(int64(1), affectedRows)

	deckFromDB, err := deckRepo.GetByID(ctx, deckID)
	suite.Require().NoError(err)
	suite.Assert().Equal(updatedDeck.Title, deckFromDB.Title)
	suite.Assert().Equal(updatedDeck.Description, deckFromDB.Description)
	suite.Assert().Equal(updatedDeck.Author, deckFromDB.Author)
}

func TestDeckTestSuite(t *testing.T) {
	suite.Run(t, new(DeckTestSuite))
}
