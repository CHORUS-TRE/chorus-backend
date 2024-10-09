package embed

import "embed"

//go:embed openapiv2/ui/*
var UIEmbed embed.FS

//go:embed openapiv2/v1-tags/*
var APIEmbed embed.FS

//go:embed dev-auth/*
var DevAuthEmbed embed.FS
