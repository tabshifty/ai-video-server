CREATE TABLE IF NOT EXISTS tv_app_releases (
    id BIGSERIAL PRIMARY KEY,
    package_name TEXT NOT NULL,
    version_code BIGINT NOT NULL,
    version_name TEXT NOT NULL DEFAULT '',
    release_notes TEXT NOT NULL DEFAULT '',
    remarks TEXT NOT NULL DEFAULT '',
    publish_status VARCHAR(32) NOT NULL DEFAULT 'draft',
    published_at TIMESTAMPTZ NULL,
    last_status_changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    latest_recommended BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT tv_app_releases_package_name_not_blank CHECK (length(trim(package_name)) > 0),
    CONSTRAINT tv_app_releases_version_code_positive CHECK (version_code > 0),
    CONSTRAINT tv_app_releases_publish_status_check CHECK (
        publish_status IN ('draft', 'published_complete', 'published_missing_abi', 'offline')
    ),
    CONSTRAINT tv_app_releases_unique_version UNIQUE (package_name, version_code)
);

CREATE INDEX IF NOT EXISTS idx_tv_app_releases_version_code
    ON tv_app_releases(package_name, version_code DESC);

CREATE INDEX IF NOT EXISTS idx_tv_app_releases_publish_status
    ON tv_app_releases(publish_status, version_code DESC);

CREATE TABLE IF NOT EXISTS tv_app_release_apks (
    id BIGSERIAL PRIMARY KEY,
    release_id BIGINT NOT NULL REFERENCES tv_app_releases(id) ON DELETE CASCADE,
    abi VARCHAR(32) NOT NULL,
    file_name TEXT NOT NULL,
    stored_path TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type TEXT NOT NULL DEFAULT 'application/vnd.android.package-archive',
    sha256 TEXT NOT NULL DEFAULT '',
    upload_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    is_debuggable BOOLEAN NOT NULL DEFAULT FALSE,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    replaced_at TIMESTAMPTZ NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    CONSTRAINT tv_app_release_apks_abi_check CHECK (abi IN ('armeabi-v7a', 'arm64-v8a')),
    CONSTRAINT tv_app_release_apks_file_size_positive CHECK (file_size > 0),
    CONSTRAINT tv_app_release_apks_file_name_not_blank CHECK (length(trim(file_name)) > 0),
    CONSTRAINT tv_app_release_apks_stored_path_not_blank CHECK (length(trim(stored_path)) > 0),
    CONSTRAINT tv_app_release_apks_release_abi_unique UNIQUE (release_id, abi)
);

CREATE INDEX IF NOT EXISTS idx_tv_app_release_apks_release_id
    ON tv_app_release_apks(release_id);
