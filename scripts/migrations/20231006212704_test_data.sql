-- +goose Up
-- +goose StatementBegin
-- Inserting test data into decks
INSERT INTO decks (id, title, description, author, created_at)
VALUES
    (1, 'Eng words', 'Description for Deck 1', 'John Doe', NOW()),
    (2, 'Deck 2', 'Description for Deck 2', 'Jane Smith', NOW());

-- Inserting test data into cards
INSERT INTO cards (front, back, deck_id, author, created_at)
VALUES
    ('Front 1', 'Back 1', 1, 'John Doe', NOW()),
    ('Front 2', 'Back 2', 1, 'Jane Smith', NOW()),
    ('Front 3', 'Back 3', 2, 'John Doe', NOW()),
    ('Front 4', 'Back 4', 2, 'Jane Smith', NOW()),
    ('Front 5', 'Back 5', 2, 'Jane Smith', NOW());

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Deleting test data from cards
DELETE FROM cards WHERE author IN ('John Doe', 'Jane Smith');

-- Deleting test data from decks
DELETE FROM decks WHERE author IN ('John Doe', 'Jane Smith');

-- +goose StatementEnd
