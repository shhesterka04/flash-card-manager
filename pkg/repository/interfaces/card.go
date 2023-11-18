//go:generate mockgen -source ./card.go -destination=./mocks/mock_card.go -package=mock_card
package interfaces

import (
	"context"
	"flash-card-manager/pkg/repository/structs"
)

type CardRepository interface {
	Add(ctx context.Context, card structs.Card) (int64, error)
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*structs.Card, error)
	Update(ctx context.Context, card structs.Card) (int64, error)
}
