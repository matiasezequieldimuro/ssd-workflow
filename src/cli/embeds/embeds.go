package embeds

import "embed"

// DefaultSDDResources embeds the default .sdd directory assets for initialization.
//go:embed all:default_sdd
var DefaultSDDResources embed.FS
