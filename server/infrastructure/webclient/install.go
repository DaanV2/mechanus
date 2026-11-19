package webclient

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
)

// MarkerFile is written into a static folder the server filled itself. It holds the hash of the client
// it was filled with, so an updated binary knows the folder is its own to replace.
const MarkerFile = ".mechanus-client"

// Install makes sure folder serves the web client, and returns the file system to serve:
//   - folder missing or empty: create it and write client into it;
//   - folder filled by an older binary (its marker holds another hash): replace it with client;
//   - folder without a marker: it is the user's (Docker image, development, custom build), serve it as is;
//   - folder can't be written: serve client straight from memory.
//
// A nil client (a binary built without the web client) serves folder as it is.
func Install(ctx context.Context, folder string, client fs.FS) http.FileSystem {
	logger := logging.FromPrefix(ctx, "webclient")
	if client == nil {
		logger.Debug("no web client embedded, serving the static folder as is", "folder", folder)

		return http.Dir(folder)
	}

	hash, err := Hash(client)
	if err != nil {
		logger.Warn("cannot read the embedded web client, serving the static folder as is", "folder", folder, "error", err)

		return http.Dir(folder)
	}

	installed, err := installedHash(folder)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		logger.Info("installing the web client", "folder", folder)
	case err != nil:
		logger.Warn("cannot read the static folder, serving the embedded web client", "folder", folder, "error", err)

		return http.FS(client)
	case installed == "":
		logger.Debug("static folder is not managed by mechanus, serving it as is", "folder", folder)

		return http.Dir(folder)
	case installed == hash:
		return http.Dir(folder)
	default:
		logger.Info("updating the web client", "folder", folder)
		if err := clearFolder(folder); err != nil {
			logger.Warn("cannot remove the old web client, serving the embedded web client", "folder", folder, "error", err)

			return http.FS(client)
		}
	}

	if err := write(folder, client, hash); err != nil {
		logger.Warn("cannot install the web client, serving the embedded web client", "folder", folder, "error", err)

		return http.FS(client)
	}

	return http.Dir(folder)
}

// installedHash returns the hash in folder's marker: fs.ErrNotExist when folder is missing or empty, empty
// when folder holds files but no marker.
func installedHash(folder string) (string, error) {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "", fs.ErrNotExist
	}

	data, err := os.ReadFile(filepath.Join(folder, MarkerFile)) // #nosec G304 -- the configured static folder
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("reading the web client marker: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}

// clearFolder removes everything in folder except the marker, and keeps folder itself, which its parent may not
// let us recreate. Keeping the (outdated) marker means an interrupted clear is redone on the next start.
func clearFolder(folder string) error {
	entries, err := os.ReadDir(folder)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.Name() == MarkerFile {
			continue
		}
		if err := os.RemoveAll(filepath.Join(folder, e.Name())); err != nil {
			return err
		}
	}

	return nil
}

// write copies client into folder. The marker is written first as "installing" and gets the hash once
// every file is in place, so a half-written folder is redone on the next start.
func write(folder string, client fs.FS, hash string) error {
	if err := os.MkdirAll(folder, 0o750); err != nil {
		return fmt.Errorf("creating the static folder: %w", err)
	}
	if err := writeMarker(folder, "installing"); err != nil {
		return err
	}

	err := fs.WalkDir(client, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(folder, filepath.FromSlash(p))
		if d.IsDir() {
			return os.MkdirAll(target, 0o750)
		}
		if p == MarkerFile || d.Name() == ".gitkeep" {
			return nil
		}

		return copyFile(client, p, target)
	})
	if err != nil {
		return fmt.Errorf("writing the web client: %w", err)
	}

	return writeMarker(folder, hash)
}

func writeMarker(folder, value string) error {
	if err := os.WriteFile(filepath.Join(folder, MarkerFile), []byte(value+"\n"), 0o600); err != nil {
		return fmt.Errorf("writing the web client marker: %w", err)
	}

	return nil
}

func copyFile(client fs.FS, p, target string) (err error) {
	src, err := client.Open(p)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, src.Close()) }()

	dst, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) // #nosec G304 -- inside the configured static folder
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, dst.Close()) }()

	_, err = io.Copy(dst, src)

	return err
}

// Hash returns a hash over the paths and contents of every file in client.
func Hash(client fs.FS) (string, error) {
	h := sha256.New()
	err := fs.WalkDir(client, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		f, err := client.Open(p)
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()

		_, _ = io.WriteString(h, p+"\x00")
		_, err = io.Copy(h, f)

		return err
	})
	if err != nil {
		return "", fmt.Errorf("hashing the web client: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
