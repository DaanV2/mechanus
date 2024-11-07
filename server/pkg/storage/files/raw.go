package file_storage

import (
	"io/fs"
	"iter"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/DaanV2/go-locks"
	xio "github.com/DaanV2/mechanus/server/pkg/extensions/io"
	"github.com/DaanV2/mechanus/server/pkg/storage"
	"github.com/charmbracelet/log"
)

type RawStorage struct {
	locks  *locks.RWPool
	folder string
}

func NewRawStorage(folder string) *RawStorage {
	xio.MakeDirAll(folder)

	l := max(runtime.GOMAXPROCS(0)*10, 10)
	return &RawStorage{
		locks:  locks.NewRWPool(l),
		folder: folder,
	}
}

func (s *RawStorage) filepath(id string) (string, *sync.RWMutex) {
	f := filepath.Join(s.folder, id) + ".dat"
	l := s.locks.GetLock(locks.KeyForString(id))

	return f, l
}

func (s *RawStorage) Set(id string, data []byte) error {
	f, l := s.filepath(id)
	l.Lock()
	defer l.Unlock()

	return os.WriteFile(f, data, 0600)
}

func (s *RawStorage) Get(id string) ([]byte, error) {
	f, l := s.filepath(id)
	l.RLock()
	defer l.RUnlock()

	data, err := os.ReadFile(f)
	if os.IsNotExist(err) {
		return nil, storage.ErrNotExist
	}

	return data, err
}

func (s *RawStorage) Has(id string) bool {
	f, l := s.filepath(id)
	l.RLock()
	defer l.RUnlock()

	fi, err := os.Stat(f)
	if os.IsNotExist(err) {
		return false
	}

	return !fi.IsDir()
}

// Ids returns an iterator over all the ids in this storage
func (s *RawStorage) Ids() iter.Seq[string] {
	return func(yield func(string) bool) {
		err := filepath.WalkDir(s.folder, func(path string, f fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if f == nil || f.IsDir() {
				return nil
			}

			name := filepath.Base(f.Name())
			ext := filepath.Ext(name)
			if len(ext) > 0 {
				name, _ = strings.CutSuffix(name, ext)
			}
			if !yield(name) {
				return filepath.SkipAll
			}

			return nil
		})

		if err != nil {
			log.Error("error attempting to read files in storage", "folder", s.folder)
		}
	}
}
