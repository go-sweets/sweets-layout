-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS data_migrations (
    id VARCHAR(255) NOT NULL PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS data_migrations;
-- +goose StatementEnd