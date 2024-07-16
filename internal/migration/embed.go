package migration

import "embed"

//go:embed postgres/*
var MigrationEmbed embed.FS
