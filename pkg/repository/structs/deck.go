package structs

import "time"

type Deck struct {
	ID          int64     `db:"id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	Author      string    `db:"author"`
	CreatedAt   time.Time `db:"created_at"`
}

type DeckWithCards struct {
	Deck  Deck
	Cards []Card `db:"cards"`
}
