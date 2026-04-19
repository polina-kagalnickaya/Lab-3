-- +goose Up
-- +goose StatementBegin
ALTER TABLE wishes ADD COLUMN user_id BIGINT;
ALTER TABLE wishes ADD CONSTRAINT fk_wishes_user_id FOREIGN KEY (user_id) REFERENCES users(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE wishes DROP CONSTRAINT fk_wishes_user_id;
ALTER TABLE wishes DROP COLUMN user_id;
-- +goose StatementEnd