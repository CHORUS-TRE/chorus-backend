package migration

import "embed"

//go:embed postgres/*
var MigrationEmbed embed.FS

//go:embed audit/postgres/*
var AuditMigrationEmbed embed.FS
