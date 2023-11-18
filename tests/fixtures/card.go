package fixtures

import (
	repository "flash-card-manager/pkg/repository/structs"
	"time"
)

type CardBuilder struct {
	instance *repository.Card
}

func Card() *CardBuilder {
	return &CardBuilder{instance: &repository.Card{}}
}

func (b *CardBuilder) ID(v int64) *CardBuilder {
	b.instance.ID = v
	return b
}

func (b *CardBuilder) Front(v string) *CardBuilder {
	b.instance.Front = v
	return b
}

func (b *CardBuilder) Back(v string) *CardBuilder {
	b.instance.Back = v
	return b
}

func (b *CardBuilder) DeckID(v int64) *CardBuilder {
	b.instance.DeckID = v
	return b
}

func (b *CardBuilder) Author(v string) *CardBuilder {
	b.instance.Author = v
	return b
}

func (b *CardBuilder) CreatedAt(v time.Time) *CardBuilder {
	b.instance.CreatedAt = v
	return b
}

func (b *CardBuilder) P() *repository.Card {
	return b.instance
}

func (b *CardBuilder) Valid() *CardBuilder {
	return Card().
		ID(6).
		Front("some front").
		Back("some back").
		DeckID(1).
		Author("Ivanov Ivan").
		CreatedAt(time.Now())
}
