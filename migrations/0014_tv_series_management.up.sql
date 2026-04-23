ALTER TABLE series
ADD COLUMN IF NOT EXISTS active BOOLEAN NOT NULL DEFAULT TRUE;

DO $$
DECLARE
    fk_name text;
BEGIN
    SELECT tc.constraint_name
    INTO fk_name
    FROM information_schema.table_constraints tc
    JOIN information_schema.key_column_usage kcu
      ON tc.constraint_name = kcu.constraint_name
     AND tc.table_schema = kcu.table_schema
    WHERE tc.table_name = 'episodes'
      AND tc.constraint_type = 'FOREIGN KEY'
      AND kcu.column_name = 'video_id'
    LIMIT 1;

    IF fk_name IS NOT NULL THEN
        EXECUTE format('ALTER TABLE episodes DROP CONSTRAINT %I', fk_name);
    END IF;
END $$;

ALTER TABLE episodes
ADD CONSTRAINT episodes_video_id_fkey
FOREIGN KEY (video_id) REFERENCES videos(id) ON DELETE SET NULL;
