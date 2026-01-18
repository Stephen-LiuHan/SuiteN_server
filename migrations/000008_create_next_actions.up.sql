-- Create next_actions table
CREATE TABLE next_actions (
    id              UUID PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ai_result_id    UUID NOT NULL REFERENCES ai_results(id) ON DELETE CASCADE,
    type            VARCHAR(20) NOT NULL,
    title           VARCHAR(200) NOT NULL,
    description     TEXT NOT NULL,
    query           VARCHAR(500) NULL,
    link            VARCHAR(1000) NULL,
    rank            INTEGER NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at      TIMESTAMP WITH TIME ZONE NULL
);

CREATE INDEX idx_next_actions_user_id ON next_actions(user_id);
CREATE INDEX idx_next_actions_ai_result_id ON next_actions(ai_result_id);
