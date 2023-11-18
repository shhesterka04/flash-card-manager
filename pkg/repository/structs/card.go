package structs

import "time"

type Card struct {
	ID        int64     `db:"id"`
	Front     string    `db:"front"`
	Back      string    `db:"back"`
	DeckID    int64     `db:"deck_id"`
	Author    string    `db:"author"`
	CreatedAt time.Time `db:"created_at"`
}
