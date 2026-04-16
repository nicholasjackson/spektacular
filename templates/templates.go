// Package templates provides embedded template files.
package templates

import "embed"

//go:embed all:*
var FS embed.FS
