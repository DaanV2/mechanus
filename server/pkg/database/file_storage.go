package database

import (
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"runtime"

	"github.com/DaanV2/go-locks"
	"github.com/DaanV2/mechanus/server/pkg/xio"
)

var _ IOHandler = &FileIO{}

// FileIO is a shareable struct that provides lock for reading / writing files
type FileIO struct {
	locks  *locks.RWPool
	folder string
}

func NewFileIO(folder string) (*FileIO, error) {
	var err error
	if !filepath.IsAbs(folder) {
		folder, err = filepath.Abs(folder)
		if err != nil {
			return nil, fmt.Errorf("couldn't create a absolute folder for storage: %w", err)
		}
	}

	xio.MakeDirAll(folder)

	return &FileIO{
		locks:  locks.NewRWPool(runtime.GOMAXPROCS(0) * 10),
		folder: folder,
	}, nil
}

func (f *FileIO) file(name string) string {
	fi := filepath.Join(f.folder, name)
	if filepath.Ext(fi) == "" {
		fi += ".dat"
	}

	xio.MakeDirAll(filepath.Dir(fi))

	return fi
}

// Get implements IOHandler.
func (f *FileIO) Get(name string) ([]byte, error) {
	file := f.file(name)
	l := f.locks.GetLock(locks.KeyForString(name))

	l.RLock()
	defer l.RUnlock()

	body, err := os.ReadFile(file)
	return body, err
}

// Set implements IOHandler.
func (f *FileIO) Set(name string, data []byte) error {
	file := f.file(name)
	l := f.locks.GetLock(locks.KeyForString(name))

	l.Lock()
	defer l.Unlock()

	return os.WriteFile(file, data, 0644)
}

func (f *FileIO) String() string {
	return "fileio: " + f.folder
}

func (f *FileIO) Ids() iter.Seq[string] {
	files, err := os.ReadDir(f.folder)
	if err != nil {
		panic(fmt.Errorf("error reading files from dir: %s -> %w", f.folder, err)) // This should never happen
	}

	return func(yield func(string) bool) {
		for _, f := range files {
			if f == nil || f.IsDir() {
				continue
			}

			name := filepath.Base(f.Name())
			if !yield(name) {
				return
			}
		}
	}
}
