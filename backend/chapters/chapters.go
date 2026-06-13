package chapters

import (
	"embed"
	"io/fs"
)

// contentFS holds every file under content/ at build time. The `all:`
// prefix includes files starting with `_` or `.`, though we don't rely
// on that — chapter files use plain names.
//
//go:embed all:content
var contentEmbed embed.FS

// contentFS is the fs.FS view registry.go reads from. Sub'd to "content"
// so registry methods can use paths like "calc/template.go".
var contentFS fs.FS

func init() {
	sub, err := fs.Sub(contentEmbed, "content")
	if err != nil {
		// Only fails if "content" doesn't exist in the embed, which is a
		// compile-time guarantee. Panic is appropriate.
		panic(err)
	}
	contentFS = sub
}
