CREATE TABLE IF NOT EXISTS auth.users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    nickname text NOT NULL,
    email text,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS auth.identities (
    provider text NOT NULL,
    provider_account_id text NOT NULL,
    user_id uuid NOT NULL REFERENCES auth.users,
    PRIMARY KEY (provider, provider_account_id)
);

CREATE TABLE IF NOT EXISTS auth.sessions (
    token_hash bytea PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES auth.users,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz
);

CREATE TABLE IF NOT EXISTS auth.login_attempts (
    state_hash bytea PRIMARY KEY,
    provider text NOT NULL,
    browser_hash bytea NOT NULL,
    verifier text NOT NULL,
    return_to text NOT NULL,
    expires_at timestamptz NOT NULL
);
