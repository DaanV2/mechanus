package util_test

import (
	"context"
	"os"
	"path/filepath"

	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/infrastructure/configfile"
	"github.com/DaanV2/mechanus/server/infrastructure/scenestore"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// WriteConfigFile writes v as the config file rel (e.g. scenes/cave.yaml) below dir, at the current version.
func WriteConfigFile(dir, rel string, v any) {
	ginkgo.GinkgoHelper()

	data, err := configfile.Encode(rel, 1, v)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	WriteRawFile(dir, rel, string(data))
}

// WriteRawFile writes content to the file rel below dir, creating folders as needed.
func WriteRawFile(dir, rel, content string) {
	ginkgo.GinkgoHelper()

	path := filepath.Join(dir, filepath.FromSlash(rel))
	gomega.Expect(os.MkdirAll(filepath.Dir(path), 0o750)).To(gomega.Succeed())
	gomega.Expect(os.WriteFile(path, []byte(content), 0o600)).To(gomega.Succeed())
}

// CreateConfigStore writes the scenes (as YAML) and the room (when not nil, as JSON) to a temporary
// config dir and returns a loaded config store on top of it, together with the dir.
func CreateConfigStore(ctx context.Context, room *devices.Room, sceneConfigs ...*scenes.Config) (store *scenestore.ConfigStore, dir string) {
	ginkgo.GinkgoHelper()

	dir = ginkgo.GinkgoT().TempDir()
	gomega.Expect(os.MkdirAll(filepath.Join(dir, "scenes"), 0o750)).To(gomega.Succeed())
	gomega.Expect(os.MkdirAll(filepath.Join(dir, "rooms"), 0o750)).To(gomega.Succeed())
	for _, cfg := range sceneConfigs {
		WriteConfigFile(dir, "scenes/"+cfg.ID+".yaml", cfg)
	}
	roomID := scenestore.RoomID
	if room != nil {
		roomID = room.ID
		WriteConfigFile(dir, "rooms/"+room.ID+".json", room)
	}

	store = scenestore.NewConfigStore(dir, roomID)
	store.Load(ctx)
	gomega.Expect(store.Errors()).To(gomega.BeEmpty())

	return store, dir
}
