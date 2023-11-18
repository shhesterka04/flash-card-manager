package fixtures

import (
	repository "flash-card-manager/pkg/repository/structs"
	"time"
)

type DeckBuilder struct {
	instance *repository.Deck
}

func Deck() *DeckBuilder {
	return &DeckBuilder{instance: &repository.Deck{}}
}

func (b *DeckBuilder) ID(v int64) *DeckBuilder {
	b.instance.ID = v
	return b
}

func (b *DeckBuilder) Title(v string) *DeckBuilder {
	b.instance.Title = v
	return b
}

func (b *DeckBuilder) Description(v string) *DeckBuilder {
	b.instance.Description = v
	return b
}

func (b *DeckBuilder) Author(v string) *DeckBuilder {
	b.instance.Author = v
	return b
}

func (b *DeckBuilder) CreatedAt(v time.Time) *DeckBuilder {
	b.instance.CreatedAt = v
	return b
}

func (b *DeckBuilder) P() *repository.Deck {
	return b.instance
}

func (b *DeckBuilder) Valid() *DeckBuilder {
	return Deck().
		ID(1).
		Title("Title").
		Description("Description").
		Author("Author").
		CreatedAt(time.Now())
}
