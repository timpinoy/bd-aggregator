-- +goose Up
CREATE TABLE feeds(
    id UUID PRIMARY KEY NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    name TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    user_id UUID NOT NULL REFERENCES users
                          ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feeds;