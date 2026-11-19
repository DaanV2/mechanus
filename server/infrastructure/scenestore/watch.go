package scenestore

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// reloadDelay groups the burst of events an editor causes when saving (truncate, write, rename, ...)
// into one reload, and avoids reading a half written file.
const reloadDelay = 150 * time.Millisecond

// Watch starts watching the scenes and rooms folders and reloads files that change, see
// [ConfigStore.Reload]. It returns once the watcher is running; [ConfigStore.StopWatching] stops it.
// Calling Watch while already watching does nothing.
func (c *ConfigStore) Watch(ctx context.Context) error {
	c.watchLock.Lock()
	defer c.watchLock.Unlock()

	if c.stopWatch != nil {
		return nil
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating config watcher: %w", err)
	}
	for _, folder := range []string{scenesFolder, roomsFolder} {
		if err := watcher.Add(filepath.Join(c.dir, folder)); err != nil {
			_ = watcher.Close()

			return fmt.Errorf("watching %s folder: %w", folder, err)
		}
	}

	ctx, cancel := context.WithCancel(context.WithoutCancel(ctx))
	c.stopWatch = cancel
	c.watchDone = make(chan struct{})
	go c.watch(ctx, watcher, c.watchDone)

	return nil
}

// StopWatching stops the watcher started by [ConfigStore.Watch] and waits for it to finish.
func (c *ConfigStore) StopWatching() {
	c.watchLock.Lock()
	defer c.watchLock.Unlock()

	if c.stopWatch == nil {
		return
	}
	c.stopWatch()
	<-c.watchDone
	c.stopWatch, c.watchDone = nil, nil
}

// AfterInitialize implements lifecycle.AfterInitialize: starts watching the config folders.
func (c *ConfigStore) AfterInitialize(ctx context.Context) error {
	return c.Watch(ctx)
}

// ShutdownCleanup implements lifecycle.ShutdownCleanup: stops watching the config folders.
func (c *ConfigStore) ShutdownCleanup(_ context.Context) error {
	c.StopWatching()

	return nil
}

func (c *ConfigStore) watch(ctx context.Context, watcher *fsnotify.Watcher, done chan struct{}) {
	logger := ConfigLogger.From(ctx)
	var (
		lock    sync.Mutex
		pending = map[string]*time.Timer{}
		reloads sync.WaitGroup
	)
	defer func() {
		_ = watcher.Close()
		lock.Lock()
		for path, t := range pending {
			if t.Stop() {
				reloads.Done()
			}
			delete(pending, path)
		}
		lock.Unlock()
		reloads.Wait()
		close(done)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			logger.Warn("config watcher error", "error", err)
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op == fsnotify.Chmod {
				continue
			}
			path := event.Name

			lock.Lock()
			if t, ok := pending[path]; ok && t.Stop() {
				t.Reset(reloadDelay)
			} else {
				reloads.Add(1)
				var timer *time.Timer
				timer = time.AfterFunc(reloadDelay, func() {
					defer reloads.Done()
					lock.Lock()
					if pending[path] == timer {
						delete(pending, path)
					}
					lock.Unlock()
					c.Reload(ctx, path)
				})
				pending[path] = timer
			}
			lock.Unlock()
		}
	}
}
