// Command embedclient copies the compiled web client into infrastructure/webclient/dist, so the next
// build of the server embeds it (see the webclient package).
//
// Usage (from server/, after `npm run build` in client/): go run ./scripts/embedclient
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

const placeholder = ".gitkeep"

func main() {
	source := flag.String("src", "../client/build", "the compiled web client")
	target := flag.String("dst", "infrastructure/webclient/dist", "the folder the webclient package embeds")
	flag.Parse()

	if _, err := os.Stat(filepath.Join(*source, "index.html")); err != nil {
		log.Fatal("no compiled web client found, run `npm run build` in client/ first", "src", *source, "error", err)
	}
	if err := clearDir(*target); err != nil {
		log.Fatal("cannot clear the embed folder", "dst", *target, "error", err)
	}
	if err := copyTree(*source, *target); err != nil {
		log.Fatal("cannot copy the web client", "src", *source, "dst", *target, "error", err)
	}

	log.Info("web client ready to embed", "src", *source, "dst", *target)
}

// clearDir removes everything in dir except the committed placeholder.
func clearDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.Name() == placeholder {
			continue
		}
		if err := os.RemoveAll(filepath.Join(dir, e.Name())); err != nil {
			return err
		}
	}

	return nil
}

func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o750)
		}

		return copyFile(p, target)
	})
}

func copyFile(src, dst string) (err error) {
	in, err := os.Open(src) // #nosec G304 -- paths come from the command line of a build script
	if err != nil {
		return fmt.Errorf("opening %s: %w", src, err)
	}
	defer func() { err = errors.Join(err, in.Close()) }()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) // #nosec G304 -- see above
	if err != nil {
		return fmt.Errorf("creating %s: %w", dst, err)
	}
	defer func() { err = errors.Join(err, out.Close()) }()

	_, err = io.Copy(out, in)

	return err
}
