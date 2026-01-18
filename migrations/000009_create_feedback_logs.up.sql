-- Create feedback_logs table
CREATE TABLE feedback_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    capture_id      UUID NOT NULL REFERENCES captures(id) ON DELETE CASCADE,
    helpfulness     INTEGER NOT NULL CHECK (helpfulness BETWEEN 1 AND 3),
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_feedback_logs_user_id ON feedback_logs(user_id);
CREATE INDEX idx_feedback_logs_capture_id ON feedback_logs(capture_id);
CREATE INDEX idx_feedback_logs_created_at ON feedback_logs(created_at);
