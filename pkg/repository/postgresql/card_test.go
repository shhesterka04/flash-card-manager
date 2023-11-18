//go:build unit
// +build unit

package postgresql

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgconn"
	"homework-3/pkg/db/mocks"
	"homework-3/pkg/repository/structs"
	"testing"
)

type mockRow struct {
	value int64
}

func (m *mockRow) Scan(dest ...interface{}) error {
	if len(dest) == 0 {
		return errors.New("no destination to scan into")
	}

	valPtr, ok := dest[0].(*int64)
	if !ok {
		return errors.New("unsupported type for Scan")
	}
	*valPtr = m.value

	return nil
}

func TestCardRepo_Add(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mock_db.NewMockDatabaseInterface(mockCtrl)
	repo := NewCard(mockDB)

	mockDB.EXPECT().ExecQueryRow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(&mockRow{value: 1})

	id, err := repo.Add(context.TODO(), structs.Card{
		Front:  "testFront",
		Back:   "testBack",
		DeckID: 1,
		Author: "testAuthor",
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if id != 1 {
		t.Errorf("Expected ID to be 1, got %d", id)
	}
}

func TestCardRepo_Delete(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mock_db.NewMockDatabaseInterface(mockCtrl)
	repo := NewCard(mockDB)

	mockDB.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil)

	err := repo.Delete(context.TODO(), 1)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCardRepo_GetByID(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mock_db.NewMockDatabaseInterface(mockCtrl)
	repo := NewCard(mockDB)

	expectedCard := structs.Card{
		ID:     1,
		Front:  "testFront",
		Back:   "testBack",
		DeckID: 1,
		Author: "testAuthor",
	}

	mockDB.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).SetArg(1, expectedCard).Return(nil)

	card, err := repo.GetByID(context.TODO(), 1)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if card.ID != 1 || card.Front != "testFront" {
		t.Errorf("Unexpected card details: %v", card)
	}
}

func TestCardRepo_Update(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockDB := mock_db.NewMockDatabaseInterface(mockCtrl)
	repo := NewCard(mockDB)

	updateCard := structs.Card{
		ID:     1,
		Front:  "updatedFront",
		Back:   "updatedBack",
		DeckID: 1,
		Author: "updatedAuthor",
	}

	mockCommandTagValue := pgconn.CommandTag("UPDATE 1")
	mockDB.EXPECT().Exec(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockCommandTagValue, nil)

	rowsAffected, err := repo.Update(context.TODO(), updateCard)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if rowsAffected != 1 {
		t.Errorf("Expected that the number of modified rows would be 1, received %d", rowsAffected)
	}
}
