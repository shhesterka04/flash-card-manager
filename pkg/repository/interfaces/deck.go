//go:generate mockgen -source ./deck.go -destination=./mocks/mock_deck.go -package=mock_deck
package interfaces

import (
	"context"
	"flash-card-manager/pkg/repository/structs"
)

type DeckRepository interface {
	Add(ctx context.Context, deck structs.Deck) (int64, error)
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*structs.Deck, error)
	Update(ctx context.Context, deck structs.Deck) (int64, error)
	GetWithCardsByID(ctx context.Context, id int64) (*structs.DeckWithCards, error)
}
