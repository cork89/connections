package main

import (
	"os"

	"github.com/evanw/esbuild/pkg/api"
)

func main() {
	result := api.Build(api.BuildOptions{
		EntryPoints:       []string{"common.ts", "create.ts", "game.ts"},
		Outdir:            "../static/",
		Bundle:            false,
		Write:             true,
		LogLevel:          api.LogLevelInfo,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Sourcemap:         api.SourceMapNone,
	})

	if len(result.Errors) > 0 {
		os.Exit(1)
	}
}
