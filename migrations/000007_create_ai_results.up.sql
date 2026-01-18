-- Create ai_results table
CREATE TABLE ai_results (
    id                  UUID PRIMARY KEY,
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    capture_id          UUID NOT NULL REFERENCES captures(id) ON DELETE CASCADE UNIQUE,
    draft_summary       TEXT NOT NULL,
    deepening_question  VARCHAR(200) NOT NULL,
    axis_selected       VARCHAR(50) NULL,
    topic_tags          JSONB DEFAULT '[]',
    created_at          TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at          TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at          TIMESTAMP WITH TIME ZONE NULL
);

CREATE INDEX idx_ai_results_user_id ON ai_results(user_id);
CREATE INDEX idx_ai_results_capture_id ON ai_results(capture_id);
CREATE INDEX idx_ai_results_updated_at ON ai_results(updated_at);
