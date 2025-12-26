-- +goose Up
-- Add OIDC identity mapping columns and unique composite index (issuer + subject)
ALTER TABLE users ADD COLUMN oidc_issuer TEXT;
ALTER TABLE users ADD COLUMN oidc_subject TEXT;
CREATE UNIQUE INDEX users_oidc_issuer_subject_key ON users(oidc_issuer, oidc_subject);

