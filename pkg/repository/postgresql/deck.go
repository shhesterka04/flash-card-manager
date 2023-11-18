package postgresql

import (
	"context"
	"database/sql"
	"errors"
	"flash-card-manager/pkg/db"
	"flash-card-manager/pkg/repository/interfaces"
	"flash-card-manager/pkg/repository/structs"
)

type DeckRepo struct {
	db db.DatabaseInterface
}

func NewDeck(database db.DatabaseInterface) interfaces.DeckRepository {
	return &DeckRepo{db: database}
}

func (r *DeckRepo) Add(ctx context.Context, deck structs.Deck) (int64, error) {
	var id int64
	err := r.db.ExecQueryRow(ctx, `INSERT INTO decks(title, description, author) VALUES($1,$2,$3) RETURNING id;`, deck.Title, deck.Description, deck.Author).Scan(&id)

	return id, err
}

func (r *DeckRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, "DELETE FROM decks WHERE id=$1", id)
	return err
}

func (r *DeckRepo) GetByID(ctx context.Context, id int64) (*structs.Deck, error) {
	var deck structs.Deck
	err := r.db.Get(ctx, &deck, "SELECT id, title, description, author, created_at FROM decks WHERE id=$1", id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("deck not found")
		}

		return nil, err
	}

	return &deck, nil
}

func (r *DeckRepo) Update(ctx context.Context, deck structs.Deck) (int64, error) {
	result, err := r.db.Exec(ctx, `UPDATE decks SET title=$1, description=$2, author=$3 WHERE id=$4;`, deck.Title, deck.Description, deck.Author, deck.ID)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

func (r *DeckRepo) GetWithCardsByID(ctx context.Context, id int64) (*structs.DeckWithCards, error) {
	query := `
	SELECT 
		d.id, 
		d.title, 
		d.description, 
		d.author, 
		d.created_at,
		c.id as "cards.id",
		c.front as "cards.front",
		c.back as "cards.back",
		c.deck_id as "cards.deck_id",
		c.author as "cards.author",
		c.created_at as "cards.created_at"
	FROM decks d
	LEFT JOIN cards c ON d.id = c.deck_id
	WHERE d.id = $1;
    `

	var rows []struct {
		structs.Deck
		Card structs.Card `db:"cards"`
	}

	err := r.db.Select(ctx, &rows, query, id)
	if err != nil {
		return nil, err
	}

	if len(rows) == 0 {
		return nil, errors.New("no deck found with the provided ID")
	}

	deckWithCards := &structs.DeckWithCards{
		Deck:  rows[0].Deck,
		Cards: []structs.Card{},
	}

	for _, row := range rows {
		if row.Card.ID != 0 {
			deckWithCards.Cards = append(deckWithCards.Cards, row.Card)
		}
	}

	return deckWithCards, nil
}
