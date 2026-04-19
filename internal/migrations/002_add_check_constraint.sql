-- +goose Up
-- +goose StatementBegin
ALTER TABLE wishes ADD CONSTRAINT wishes_priority_check 
    CHECK (priority >= 1 AND priority <= 5);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE wishes DROP CONSTRAINT IF EXISTS wishes_priority_check;
-- +goose StatementEnd