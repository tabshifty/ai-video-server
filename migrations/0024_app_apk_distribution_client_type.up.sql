ALTER TABLE tv_app_releases
    ADD COLUMN IF NOT EXISTS client_type VARCHAR(32) NOT NULL DEFAULT 'android_tv';

UPDATE tv_app_releases
SET client_type = 'android_tv'
WHERE client_type IS NULL OR length(trim(client_type)) = 0;

ALTER TABLE tv_app_releases
    DROP CONSTRAINT IF EXISTS tv_app_releases_client_type_check;

ALTER TABLE tv_app_releases
    ADD CONSTRAINT tv_app_releases_client_type_check CHECK (
        client_type IN ('android_tv', 'android_phone')
    );

ALTER TABLE tv_app_releases
    DROP CONSTRAINT IF EXISTS tv_app_releases_client_type_package_name_check;

ALTER TABLE tv_app_releases
    ADD CONSTRAINT tv_app_releases_client_type_package_name_check CHECK (
        (client_type = 'android_tv' AND package_name = 'com.chee.videos.tv')
        OR
        (client_type = 'android_phone' AND package_name = 'com.chee.videos')
    );

CREATE INDEX IF NOT EXISTS idx_tv_app_releases_client_type_version_code
    ON tv_app_releases(client_type, version_code DESC, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_tv_app_releases_client_type_publish_status
    ON tv_app_releases(client_type, publish_status, version_code DESC);

ALTER TABLE tv_app_release_apks
    DROP CONSTRAINT IF EXISTS tv_app_release_apks_abi_check;

ALTER TABLE tv_app_release_apks
    ADD CONSTRAINT tv_app_release_apks_abi_check CHECK (
        abi IN ('armeabi-v7a', 'arm64-v8a', 'single')
    );
