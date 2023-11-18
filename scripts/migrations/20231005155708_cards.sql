-- +goose Up
-- +goose StatementBegin
CREATE TABLE cards(
                      id SERIAL PRIMARY KEY,
                      front TEXT NOT NULL,
                      back TEXT NOT NULL,
                      deck_id INT NOT NULL,
                      author TEXT NOT NULL,
                      created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE cards;
-- +goose StatementEnd
