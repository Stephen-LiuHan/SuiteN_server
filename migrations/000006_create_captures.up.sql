-- Create captures table
CREATE TABLE captures (
    id              UUID PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id      UUID NOT NULL REFERENCES reading_sessions(id) ON DELETE CASCADE,
    emotion         VARCHAR(10) NULL,
    user_memo       VARCHAR(500) NULL,
    ai_status       VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at      TIMESTAMP WITH TIME ZONE NULL
);

CREATE INDEX idx_captures_user_id ON captures(user_id);
CREATE INDEX idx_captures_session_id ON captures(session_id);
CREATE INDEX idx_captures_updated_at ON captures(updated_at);
