package staticsrc

import "embed"

// Content is the static source data.
//
//go:embed *.json
var Content embed.FS
