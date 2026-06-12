DELETE FROM tv_app_releases
WHERE client_type = 'android_phone';

DROP INDEX IF EXISTS idx_tv_app_releases_client_type_publish_status;
DROP INDEX IF EXISTS idx_tv_app_releases_client_type_version_code;

ALTER TABLE tv_app_release_apks
    DROP CONSTRAINT IF EXISTS tv_app_release_apks_abi_check;

ALTER TABLE tv_app_release_apks
    ADD CONSTRAINT tv_app_release_apks_abi_check CHECK (
        abi IN ('armeabi-v7a', 'arm64-v8a')
    );

ALTER TABLE tv_app_releases
    DROP CONSTRAINT IF EXISTS tv_app_releases_client_type_package_name_check;

ALTER TABLE tv_app_releases
    DROP CONSTRAINT IF EXISTS tv_app_releases_client_type_check;

ALTER TABLE tv_app_releases
    DROP COLUMN IF EXISTS client_type;
