#!/bin/sh
set -eu

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" \
    -v auth_password="$AUTH_DB_PASSWORD" -v library_password="$LIBRARY_DB_PASSWORD" <<'SQL'
CREATE ROLE frets_auth LOGIN PASSWORD :'auth_password';
CREATE ROLE frets_library LOGIN PASSWORD :'library_password';
REVOKE CREATE ON SCHEMA public FROM PUBLIC;

CREATE SCHEMA auth AUTHORIZATION frets_auth;
CREATE SCHEMA song_library AUTHORIZATION frets_library;
SQL
