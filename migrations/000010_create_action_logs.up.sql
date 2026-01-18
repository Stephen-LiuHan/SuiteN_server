-- Create action_logs table
CREATE TABLE action_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ai_result_id    UUID NOT NULL REFERENCES ai_results(id) ON DELETE CASCADE,
    action_type     VARCHAR(50) NOT NULL,
    opened          BOOLEAN DEFAULT FALSE,
    completed       BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_action_logs_user_id ON action_logs(user_id);
CREATE INDEX idx_action_logs_ai_result_id ON action_logs(ai_result_id);
CREATE INDEX idx_action_logs_created_at ON action_logs(created_at);
