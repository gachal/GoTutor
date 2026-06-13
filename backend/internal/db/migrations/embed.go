// Package migrations embeds the SQL migration files so they ship inside
// the Go binary and are applied automatically on first DB connect.
package migrations

import (
	"embed"
	"io/fs"
	"sort"
)

//go:embed *.sql
var migrationFS embed.FS

// FS returns the embedded migrations as a fs.FS.
func FS() fs.FS {
	sub, err := fs.Sub(migrationFS, ".")
	if err != nil {
		panic(err)
	}
	return sub
}

// SortedNames returns migration filenames in lexical order so db.Open
// can apply them deterministically (0001_init.sql before 0002_*, etc.).
func SortedNames() []string {
	entries, err := fs.ReadDir(migrationFS, ".")
	if err != nil {
		panic(err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names
}
