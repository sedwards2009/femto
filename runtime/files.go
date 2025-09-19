//go:generate go run assets_generate.go

package runtime

import "github.com/sedwards2009/femto"

var Files = femto.NewRuntimeFiles(files)
