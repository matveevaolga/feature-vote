CREATE TABLE IF NOT EXISTS votings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id),
    feature_name VARCHAR(200) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'cancelled')),
    duration BIGINT NOT NULL,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    ends_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,
    result BOOLEAN,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_votings_group_id ON votings(group_id);
CREATE INDEX idx_votings_status ON votings(status);
CREATE INDEX idx_votings_ends_at ON votings(ends_at);

CREATE TABLE IF NOT EXISTS votes (
    voting_id UUID NOT NULL REFERENCES votings(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vote VARCHAR(3) NOT NULL CHECK (vote IN ('yes', 'no')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (voting_id, user_id)
);

CREATE INDEX idx_votes_voting_id ON votes(voting_id);