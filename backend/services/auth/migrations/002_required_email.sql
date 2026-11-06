-- Preserve legacy profiles until their next sign-in; enforce email for every new write.
ALTER TABLE auth.users
    ADD CONSTRAINT users_email_required
    CHECK (email IS NOT NULL AND btrim(email) <> '') NOT VALID;
