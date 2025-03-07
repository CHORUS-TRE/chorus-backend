-- +migrate Up

-- +migrate StatementBegin
ALTER TABLE app_instances RENAME COLUMN initial_resolution_width TO initialresolutionwidth;
ALTER TABLE app_instances RENAME COLUMN initial_resolution_height TO initialresolutionheight;
-- +migrate StatementEnd
