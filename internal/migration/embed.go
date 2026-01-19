package migration

import "embed"

//go:embed chorus/postgres/*
var ChorusMigrationEmbed embed.FS

//go:embed audit/postgres/*
var AuditMigrationEmbed embed.FS
