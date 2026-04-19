-- +goose Up
-- +goose StatementBegin
CREATE TABLE wishes (
    id BIGSERIAL PRIMARY KEY,
    text TEXT NOT NULL,
    author VARCHAR(100) DEFAULT 'Anonymous',
    priority INTEGER DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_wishes_deleted_at ON wishes(deleted_at);
CREATE INDEX idx_wishes_priority ON wishes(priority);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS wishes;
-- +goose StatementEnd