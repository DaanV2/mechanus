package database

import (
	"fmt"
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
