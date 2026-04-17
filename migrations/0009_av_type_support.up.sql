ALTER TABLE videos
    DROP CONSTRAINT IF EXISTS videos_type_check;

ALTER TABLE videos
    ADD CONSTRAINT videos_type_check
        CHECK (type IN ('short','movie','episode','av'));
