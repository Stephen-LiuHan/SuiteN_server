-- Create user_profiles table
CREATE TABLE user_profiles (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                 UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    axis_preferences        JSONB DEFAULT '{"risk":0,"success_rate":0,"rationale":0,"concept":0}',
    action_preferences      JSONB DEFAULT '{"dialog":0,"books":0,"videos":0}',
    updated_at              TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_user_profiles_user_id ON user_profiles(user_id);
