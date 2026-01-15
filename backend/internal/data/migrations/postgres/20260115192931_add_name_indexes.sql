-- +goose Up
CREATE INDEX IF NOT EXISTS idx_locations_name ON public.locations(name);
CREATE INDEX IF NOT EXISTS idx_labels_name ON public.labels(name);

-- +goose Down
DROP INDEX IF EXISTS public.idx_locations_name;
DROP INDEX IF EXISTS public.idx_labels_name;
