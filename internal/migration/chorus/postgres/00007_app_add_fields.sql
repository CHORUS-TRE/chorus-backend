-- +migrate Up

-- +migrate StatementBegin
ALTER TABLE apps ADD COLUMN shmsize TEXT NOT NULL DEFAULT '';
ALTER TABLE apps ADD COLUMN kioskconfigurl TEXT NOT NULL DEFAULT '';
ALTER TABLE apps ADD COLUMN maxcpu TEXT NOT NULL DEFAULT '';
ALTER TABLE apps ADD COLUMN mincpu TEXT NOT NULL DEFAULT '';
ALTER TABLE apps ADD COLUMN maxmemory TEXT NOT NULL DEFAULT '';
ALTER TABLE apps ADD COLUMN minmemory TEXT NOT NULL DEFAULT '';
ALTER TABLE apps ADD COLUMN iconurl TEXT NOT NULL DEFAULT '';
-- +migrate StatementEnd

-- +migrate StatementBegin
ALTER TABLE app_instances ADD COLUMN initial_resolution_width INT NOT NULL DEFAULT 0;
ALTER TABLE app_instances ADD COLUMN initial_resolution_height INT NOT NULL DEFAULT 0;
-- +migrate StatementEnd
