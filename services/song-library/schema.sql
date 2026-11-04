CREATE TABLE IF NOT EXISTS song_library.songs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id uuid NOT NULL,
    title text,
    created_at timestamptz NOT NULL DEFAULT now(),
    status text NOT NULL CHECK (status IN ('uploading','ready','deleting','deleted')),
    upload_key text,
    sha256 text,
    size_bytes bigint,
    file_key text,
    UNIQUE (owner_id, upload_key),
    CHECK ((status IN ('ready','deleting')) = (file_key IS NOT NULL))
);

CREATE TABLE IF NOT EXISTS song_library.upload_attempts (
    id uuid PRIMARY KEY,
    song_id uuid NOT NULL REFERENCES song_library.songs,
    file_key text NOT NULL UNIQUE,
    status text NOT NULL CHECK (status IN ('active','expired','published')),
    expires_at timestamptz NOT NULL,
    sweep_after timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS upload_active ON song_library.upload_attempts(song_id)
    WHERE status = 'active';
CREATE INDEX IF NOT EXISTS upload_sweep ON song_library.upload_attempts(sweep_after)
    WHERE status = 'expired';

CREATE TABLE IF NOT EXISTS song_library.fragments (
    id uuid PRIMARY KEY,
    song_id uuid NOT NULL REFERENCES song_library.songs,
    title text NOT NULL,
    start_bar integer NOT NULL,
    end_bar integer NOT NULL,
    CHECK (1 <= start_bar AND start_bar <= end_bar)
);

CREATE INDEX IF NOT EXISTS fragments_song_id ON song_library.fragments(song_id);
