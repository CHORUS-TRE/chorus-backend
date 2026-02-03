-- +migrate Up

-- +migrate StatementBegin
ALTER TABLE workbenchs ADD COLUMN initialresolutionwidth INT NOT NULL DEFAULT 0;
ALTER TABLE workbenchs ADD COLUMN initialresolutionheight INT NOT NULL DEFAULT 0;
-- +migrate StatementEnd