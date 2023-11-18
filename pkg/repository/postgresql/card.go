package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"flash-card-manager/pkg/db"
	"flash-card-manager/pkg/repository/interfaces"
	"flash-card-manager/pkg/repository/structs"
)

type CardRepo struct {
	db db.DatabaseInterface
}

func NewCard(database db.DatabaseInterface) interfaces.CardRepository {
	return &CardRepo{db: database}
}

func (r *CardRepo) Add(ctx context.Context, card structs.Card) (int64, error) {
	var id int64
	err := r.db.ExecQueryRow(ctx, `INSERT INTO cards(front, back, deck_id, author) VALUES($1,$2,$3,$4) RETURNING id;`, card.Front, card.Back, card.DeckID, card.Author).Scan(&id)

	return id, err
}

func (r *CardRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, "DELETE FROM cards WHERE id=$1", id)
	return err
}

func (r *CardRepo) GetByID(ctx context.Context, id int64) (*structs.Card, error) {
	var card structs.Card
	err := r.db.Get(ctx, &card, "SELECT id, front, back, deck_id, author, created_at FROM cards WHERE id=$1", id)

	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {

			return nil, errors.New("card not found")
		}

		return nil, err
	}

	return &card, nil
}

func (r *CardRepo) Update(ctx context.Context, card structs.Card) (int64, error) {
	result, err := r.db.Exec(ctx, `UPDATE cards SET front=$1, back=$2, deck_id=$3, author=$4 WHERE id=$5;`, card.Front, card.Back, card.DeckID, card.Author, card.ID)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}
