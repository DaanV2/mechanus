package storage

import (
	"context"
	"iter"
	"os"
	"path/filepath"

	"github.com/DaanV2/mechanus/server/infrastructure/logging"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xformat"
	"github.com/DaanV2/mechanus/server/pkg/extensions/xio"
	"github.com/DaanV2/mechanus/server/pkg/paths"
	"github.com/daanv2/go-kit/generics"
	"github.com/daanv2/go-locks"
)

func FileStorage[T Identifiable]() StorageProvider[T] {
	return &fileStorage[T]{
		locks: *locks.NewRWPool(),
	}
}

type fileStorage[T Identifiable] struct {
	locks locks.RWPool
}

type fileDirStorage[T Identifiable] struct {
	base *fileStorage[T]
	dir  string
}

// AppStorage implements StorageProvider.
func (f *fileStorage[T]) AppStorage() (Storage[T], error) {
	dir, err := paths.GetAppConfigDir()
	if err != nil {
		return nil, err
	}

	dir = filepath.Join(dir, generics.NameOf[T]())
	xio.MakeDirAll(dir)

	return &fileDirStorage[T]{
		base: f,
		dir:  dir,
	}, nil
}

// StateStorage implements StorageProvider.
func (f *fileStorage[T]) StateStorage() (Storage[T], error) {
	dir, err := paths.GetStateDir()
	if err != nil {
		return nil, err
	}

	dir = filepath.Join(dir, generics.NameOf[T]())
	xio.MakeDirAll(dir)

	return &fileDirStorage[T]{
		base: f,
		dir:  dir,
	}, nil
}

// UserStorage implements StorageProvider.
func (f *fileStorage[T]) UserStorage() (Storage[T], error) {
	dir, err := paths.GetUserDataDir()
	if err != nil {
		return nil, err
	}

	dir = filepath.Join(dir, generics.NameOf[T]())
	xio.MakeDirAll(dir)

	return &fileDirStorage[T]{
		base: f,
		dir:  dir,
	}, nil
}

// Delete implements Storage.
func (f *fileDirStorage[T]) Delete(ctx context.Context, item T) (bool, error) {
	path := f.filepath(item.GetID())
	logging.FromPrefix(ctx, "file-storage").Debug("deleting file: " + path)

	l := f.base.locks.GetLockByString(path)
	l.Lock()
	defer l.Unlock()

	err := os.Remove(path)

	return err != nil, err
}

// Get implements Storage.
func (f *fileDirStorage[T]) Get(ctx context.Context, id string) (T, error) {
	path := f.filepath(id)

	d, err := f.read(ctx, path)
	if err != nil {
		return generics.Empty[T](), err
	}

	return unmarshallGeneric[T](d)
}

// Set implements Storage.
func (f *fileDirStorage[T]) Set(ctx context.Context, item T) error {
	d, err := xformat.Marshal(item)
	if err != nil {
		return err
	}

	path := f.filepath(item.GetID())

	return f.write(ctx, path, d)
}

func (f *fileDirStorage[T]) Keys(ctx context.Context) iter.Seq[string] {
	return func(yield func(string) bool) {
		entries, err := os.ReadDir(f.dir)
		if err != nil {
			return
		}
		for _, entry := range entries {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			ext := filepath.Ext(name)
			if ext != ".dat" {
				continue
			}
			id := name[:len(name)-len(ext)]
			if !yield(id) {
				return
			}
		}
	}
}

func (f *fileDirStorage[T]) filepath(id string) string {
	return filepath.Join(f.dir, id, ".dat")
}

// Get implements Storage.
func (f *fileDirStorage[T]) read(ctx context.Context, path string) ([]byte, error) {
	logging.FromPrefix(ctx, "file-storage").Debug("reading file: " + path)

	l := f.base.locks.GetLockByString(path)
	l.RLock()
	defer l.RUnlock()

	//nolint:gosec //file checking should be done higher up the chain
	return os.ReadFile(path)
}

// Get implements Storage.
func (f *fileDirStorage[T]) write(ctx context.Context, path string, data []byte) error {
	logging.FromPrefix(ctx, "file-storage").Debug("writing file: "+path, "len", len(data))

	l := f.base.locks.GetLockByString(path)
	l.Lock()
	defer l.Unlock()

	return os.WriteFile(path, data, xio.DEFAULT_FILE_PERMISSIONS)
}
