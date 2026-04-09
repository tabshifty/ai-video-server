CREATE TABLE IF NOT EXISTS file_hashes (
    hash VARCHAR(64) PRIMARY KEY,
    video_id UUID NOT NULL REFERENCES videos(id) ON DELETE CASCADE,
    file_size BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_file_hashes_hash ON file_hashes(hash);
CREATE INDEX IF NOT EXISTS idx_file_hashes_video_id ON file_hashes(video_id);

