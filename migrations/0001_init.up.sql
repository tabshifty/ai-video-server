CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS videos (
    id UUID PRIMARY KEY,
    user_id UUID,
    title VARCHAR(200),
    description TEXT,
    type VARCHAR(10) NOT NULL CHECK (type IN ('short','movie','episode')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('uploaded','processing','ready','failed')),
    duration_seconds INT,
    width INT,
    height INT,
    original_path TEXT,
    transcoded_path TEXT,
    thumbnail_path TEXT,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_videos_type_status ON videos(type, status);
CREATE INDEX IF NOT EXISTS idx_videos_created ON videos(created_at DESC);

CREATE TABLE IF NOT EXISTS video_tags (
    video_id UUID REFERENCES videos(id) ON DELETE CASCADE,
    tag VARCHAR(50) NOT NULL,
    weight DOUBLE PRECISION DEFAULT 1.0,
    PRIMARY KEY (video_id, tag)
);
CREATE INDEX IF NOT EXISTS idx_tag_videos ON video_tags(tag);

CREATE TABLE IF NOT EXISTS user_video_actions (
    id BIGSERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    video_id UUID NOT NULL REFERENCES videos(id),
    action_type VARCHAR(20) NOT NULL,
    watch_seconds INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, video_id, action_type)
);
CREATE INDEX IF NOT EXISTS idx_user_actions ON user_video_actions(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS series (
    id SERIAL PRIMARY KEY,
    tmdb_id INT UNIQUE,
    title VARCHAR(255) NOT NULL,
    overview TEXT,
    poster_path VARCHAR(255),
    backdrop_path VARCHAR(255),
    first_air_date DATE,
    number_of_seasons INT,
    number_of_episodes INT
);

CREATE TABLE IF NOT EXISTS seasons (
    id SERIAL PRIMARY KEY,
    series_id INT NOT NULL REFERENCES series(id) ON DELETE CASCADE,
    season_number INT NOT NULL,
    name VARCHAR(255),
    overview TEXT,
    poster_path VARCHAR(255),
    air_date DATE,
    UNIQUE(series_id, season_number)
);

CREATE TABLE IF NOT EXISTS episodes (
    id SERIAL PRIMARY KEY,
    season_id INT NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
    episode_number INT NOT NULL,
    title VARCHAR(255),
    overview TEXT,
    still_path VARCHAR(255),
    runtime INT,
    air_date DATE,
    video_id UUID UNIQUE REFERENCES videos(id) ON DELETE CASCADE,
    UNIQUE(season_id, episode_number)
);

CREATE TABLE IF NOT EXISTS transcoding_jobs (
    id BIGSERIAL PRIMARY KEY,
    video_id UUID REFERENCES videos(id),
    status VARCHAR(20),
    retry_count INT DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ
);
