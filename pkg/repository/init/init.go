package repository

import (
	"flash-card-manager/pkg/db"
	"flash-card-manager/pkg/repository/interfaces"
	"flash-card-manager/pkg/repository/postgresql"
)

func InitRepositories(database db.DatabaseInterface) (interfaces.CardRepository, interfaces.DeckRepository) {
	cardRepo := postgresql.NewCard(database)
	deckRepo := postgresql.NewDeck(database)
	return cardRepo, deckRepo
}
