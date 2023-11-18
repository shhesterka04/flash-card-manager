-- +goose Up
-- +goose StatementBegin
ALTER TABLE cards
    ADD CONSTRAINT fk_deck_id
        FOREIGN KEY (deck_id)
            REFERENCES decks(id)
            ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE cards
    DROP CONSTRAINT fk_deck_id;
-- +goose StatementEnd
