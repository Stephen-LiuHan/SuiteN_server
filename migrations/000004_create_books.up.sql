-- Create books table
CREATE TABLE books (
    id              UUID PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title           VARCHAR(500) NOT NULL,
    author          VARCHAR(255) NULL,
    summary         TEXT NULL,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at      TIMESTAMP WITH TIME ZONE NULL
);

CREATE INDEX idx_books_user_id ON books(user_id);
CREATE INDEX idx_books_updated_at ON books(updated_at);
CREATE INDEX idx_books_user_updated ON books(user_id, updated_at);
