//go:build unit
// +build unit

package postgresql

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jackc/pgconn"
	"homework-3/pkg/db/mocks"
	"homework-3/pkg/repository/structs"
)

func TestDeckRepo_Add(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mock_db.NewMockDatabaseInterface(mockCtrl)
	repo := NewDeck(mockDB)

	mockDB.EXPECT().ExecQueryRow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&mockRow{value: 1})

	id, err := repo.Add(context.TODO(), structs.Deck{
		Title:       "testTitle",
		Description: "testDescription",
		Author:      "testAuthor",
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if id != 1 {
		t.Errorf("Expected ID to be 1, got %d", id)
	}
}

func TestDeckRepo_Delete(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mock_db.NewMockDatabaseInterface(mockCtrl)
	repo := NewDeck(mockDB)

	mockDB.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)

	err := repo.Delete(context.TODO(), 1)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestDeckRepo_GetByID(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mock_db.NewMockDatabaseInterface(mockCtrl)
	repo := NewDeck(mockDB)

	expectedDeck := structs.Deck{
		ID:          1,
		Title:       "testTitle",
		Description: "testDescription",
		Author:      "testAuthor",
	}

	mockDB.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).SetArg(1, expectedDeck).Return(nil)

	deck, err := repo.GetByID(context.TODO(), 1)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if deck.ID != 1 || deck.Title != "testTitle" {
		t.Errorf("Unexpected deck details: %v", deck)
	}
}

func TestDeckRepo_Update(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mock_db.NewMockDatabaseInterface(mockCtrl)
	repo := NewDeck(mockDB)

	updateDeck := structs.Deck{
		ID:          1,
		Title:       "updatedTitle",
		Description: "updatedDescription",
		Author:      "updatedAuthor",
	}

	mockCommandTagValue := pgconn.CommandTag("UPDATE 1")
	mockDB.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockCommandTagValue, nil)

	rowsAffected, err := repo.Update(context.TODO(), updateDeck)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if rowsAffected != 1 {
		t.Errorf("Expected rows affected to be 1, got %d", rowsAffected)
	}
}

func TestDeckRepo_GetWithCardsByID(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mock_db.NewMockDatabaseInterface(mockCtrl)
	repo := NewDeck(mockDB)

	expectedDeckWithCards := structs.DeckWithCards{
		Deck: structs.Deck{
			ID:          1,
			Title:       "testTitle",
			Description: "testDescription",
			Author:      "testAuthor",
		},
		Cards: []structs.Card{
			{
				ID:     1,
				Front:  "cardFront1",
				Back:   "cardBack1",
				DeckID: 1,
				Author: "cardAuthor1",
			},
		},
	}

	rows := []struct {
		structs.Deck
		Card structs.Card `db:"cards"`
	}{
		{
			Deck: expectedDeckWithCards.Deck,
			Card: expectedDeckWithCards.Cards[0],
		},
	}

	mockDB.EXPECT().Select(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).SetArg(1, rows).Return(nil)

	deckWithCards, err := repo.GetWithCardsByID(context.TODO(), 1)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if deckWithCards.Deck.ID != 1 || deckWithCards.Deck.Title != "testTitle" {
		t.Errorf("Unexpected deck details: %v", deckWithCards.Deck)
	}
	if len(deckWithCards.Cards) != 1 || deckWithCards.Cards[0].Front != "cardFront1" {
		t.Errorf("Unexpected cards in deck: %v", deckWithCards.Cards)
	}
}
