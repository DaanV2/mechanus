package webclient

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// Embedded returns the web client compiled into the binary; ok is false when the binary was built without
// one (dist/ only holds its placeholder).
func Embedded() (client fs.FS, ok bool) {
	client, err := fs.Sub(dist, "dist")
	if err != nil {
		return nil, false
	}
	if _, err := fs.Stat(client, "index.html"); err != nil {
		return nil, false
	}

	return client, true
}
