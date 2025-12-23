-- +goose Up
-- Add OIDC identity mapping columns and unique composite index (issuer + subject)
ALTER TABLE public.users ADD COLUMN oidc_issuer VARCHAR;
ALTER TABLE public.users ADD COLUMN oidc_subject VARCHAR;
-- Partial unique index so multiple NULL pairs are allowed, enforcing uniqueness only when both present.
CREATE UNIQUE INDEX users_oidc_issuer_subject_key ON public.users(oidc_issuer, oidc_subject)
  WHERE oidc_issuer IS NOT NULL AND oidc_subject IS NOT NULL;

