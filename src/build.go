package main

import (
	"fmt"
	"os"

	"github.com/evanw/esbuild/pkg/api"
)

func main() {
	result := api.Build(api.BuildOptions{
		EntryPoints:       []string{"./src/common.ts", "./src/create.ts", "./src/game.ts", "./src/mygames.ts"},
		Outdir:            "./static/",
		Bundle:            false,
		Write:             true,
		LogLevel:          api.LogLevelInfo,
		MinifyWhitespace:  true,
		MinifyIdentifiers: true,
		MinifySyntax:      true,
		Sourcemap:         api.SourceMapNone,
	})

	if len(result.Errors) > 0 {
		fmt.Println(result.Errors)
		os.Exit(1)
	}
}
