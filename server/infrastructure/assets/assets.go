package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/DaanV2/mechanus/server/pkg/paths"
)

const (
	// SharedPrefix marks an asset reference as pointing into the shared assets folder.
	SharedPrefix = "shared:/"
	// RoutePrefix is the URL path assets are served under.
	RoutePrefix = "/assets/"
	// VersionQuery is the query parameter holding the content hash.
	VersionQuery = "v"

	scenesDir = "scenes"
	sharedDir = "shared"
)

var (
	// ErrInvalidAsset is returned for references that point outside the asset folders.
	ErrInvalidAsset = errors.New("invalid asset reference")
)

// Resolved is an asset reference resolved to something a screen can load.
type Resolved struct {
	URL       string // URL the client loads the asset from
	MediaType string // MIME type guessed from the extension, empty when unknown
}

// Store resolves asset references and serves the asset files. Safe for concurrent use.
type Store struct {
	root fs.FS

	lock   sync.Mutex
	hashes map[string]hashEntry // file path -> content hash, invalidated on size / modtime change
}

type hashEntry struct {
	size    int64
	modTime time.Time
	hash    string
}

// NewStore creates a store serving assets from the given file system, laid out as described in the
// package documentation.
func NewStore(root fs.FS) *Store {
	return &Store{root: root, hashes: map[string]hashEntry{}}
}

// NewAppStore creates a store serving assets from the app config folder.
func NewAppStore() (*Store, error) {
	dir, err := paths.GetAppConfigDir()
	if err != nil {
		return nil, fmt.Errorf("asset folder: %w", err)
	}

	return NewStore(os.DirFS(dir)), nil
}

// Resolve turns an asset reference from the configuration of the given scene into a URL with a
// content hash. Absolute http(s) URLs are passed through untouched.
func (s *Store) Resolve(sceneID, ref string) (*Resolved, error) {
	if ref == "" {
		return nil, fmt.Errorf("empty reference: %w", ErrInvalidAsset)
	}
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		return &Resolved{URL: ref, MediaType: mediaType(ref)}, nil
	}

	p, err := filePath(sceneID, ref)
	if err != nil {
		return nil, err
	}
	hash, err := s.hash(p)
	if err != nil {
		return nil, fmt.Errorf("asset %q of scene %q: %w", ref, sceneID, err)
	}

	u := url.URL{Path: RoutePrefix + p, RawQuery: url.Values{VersionQuery: {hash}}.Encode()}

	return &Resolved{URL: u.String(), MediaType: mediaType(p)}, nil
}

// filePath maps a reference to a path in the asset file system.
func filePath(sceneID, ref string) (string, error) {
	var p string
	if rest, ok := strings.CutPrefix(ref, SharedPrefix); ok {
		p = path.Join(sharedDir, rest)
	} else {
		if sceneID == "" || !fs.ValidPath(sceneID) || strings.Contains(sceneID, "/") {
			return "", fmt.Errorf("scene %q: %w", sceneID, ErrInvalidAsset)
		}
		p = path.Join(scenesDir, sceneID, ref)
	}
	if !servable(p) {
		return "", fmt.Errorf("%q: %w", ref, ErrInvalidAsset)
	}

	return p, nil
}

// servable reports whether a file path may be served: only files inside a scene's folder or the
// shared folder. This keeps scene configs (scenes/<id>.dat) and everything else private.
func servable(p string) bool {
	if !fs.ValidPath(p) {
		return false
	}
	parts := strings.Split(p, "/")
	switch parts[0] {
	case sharedDir:
		return len(parts) >= 2
	case scenesDir:
		return len(parts) >= 3
	default:
		return false
	}
}

func (s *Store) hash(p string) (string, error) {
	info, err := fs.Stat(s.root, p)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("%q is a folder: %w", p, ErrInvalidAsset)
	}

	s.lock.Lock()
	entry, ok := s.hashes[p]
	s.lock.Unlock()
	if ok && entry.size == info.Size() && entry.modTime.Equal(info.ModTime()) {
		return entry.hash, nil
	}

	f, err := s.root.Open(p)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	sum := hex.EncodeToString(h.Sum(nil))[:16]

	s.lock.Lock()
	s.hashes[p] = hashEntry{size: info.Size(), modTime: info.ModTime(), hash: sum}
	s.lock.Unlock()

	return sum, nil
}

// ServeHTTP serves asset files below [RoutePrefix]. Requests carrying a version hash are cacheable
// forever, as a changed file gets a new hash.
func (s *Store) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p, ok := strings.CutPrefix(r.URL.Path, RoutePrefix)
	if !ok || !servable(p) {
		http.NotFound(w, r)

		return
	}

	info, err := fs.Stat(s.root, p)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)

		return
	}

	if r.URL.Query().Get(VersionQuery) != "" {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}

	http.ServeFileFS(w, r, s.root, p) //nolint:gosec // G703 p is checked by servable: a valid fs path inside the asset folders
}

func mediaType(p string) string {
	if u, err := url.Parse(p); err == nil {
		p = u.Path
	}

	return mime.TypeByExtension(path.Ext(p))
}
