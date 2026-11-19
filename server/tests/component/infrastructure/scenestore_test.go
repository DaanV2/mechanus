package infrastructure_test

import (
	"context"

	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/infrastructure/scenestore"
	"github.com/DaanV2/mechanus/server/infrastructure/storage"
	util_test "github.com/DaanV2/mechanus/server/tests/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	_ application.SceneConfigStore = (*scenestore.ConfigStore)(nil)
	_ application.RoomConfigStore  = (*scenestore.ConfigStore)(nil)
)

var _ = Describe("scenestore ConfigStore", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	Describe("ConfigStore", func() {
		var (
			store        *scenestore.ConfigStore
			sceneStorage storage.Storage[*scenes.Config]
			roomStorage  storage.Storage[*devices.Room]
		)

		// The adapter logic is storage agnostic; the database stands in for files to avoid touching the user's config dir.
		BeforeEach(func() {
			db := util_test.CreateDatabase(ctx)
			var err error
			sceneStorage, err = storage.NamedDBStorage[*scenes.Config](db, "scenes").AppStorage()
			Expect(err).ToNot(HaveOccurred())
			roomStorage, err = storage.NamedDBStorage[*devices.Room](db, "rooms").AppStorage()
			Expect(err).ToNot(HaveOccurred())

			Expect(sceneStorage.Set(ctx, &scenes.Config{ID: "cave", Name: "Cave"})).To(Succeed())
			Expect(roomStorage.Set(ctx, devices.NewRoom("living-room", "Living room"))).To(Succeed())
			store = scenestore.NewConfigStore(sceneStorage, roomStorage, "living-room")
		})

		It("reads scene configs", func() {
			cfg, err := store.GetSceneConfig(ctx, "cave")
			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.Name).To(Equal("Cave"))
		})

		It("reports missing scenes", func() {
			_, err := store.GetSceneConfig(ctx, "road")
			Expect(err).To(MatchError(scenes.ErrSceneNotFound))
		})

		It("rejects ids that are unsafe as file names", func() {
			_, err := store.GetSceneConfig(ctx, "../secrets")
			Expect(err).To(MatchError(scenestore.ErrInvalidID))
		})

		It("reads the current room", func() {
			room, err := store.CurrentRoom(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(room.Name).To(Equal("Living room"))
		})

		It("lists scene configs sorted by id", func() {
			Expect(sceneStorage.Set(ctx, &scenes.Config{ID: "attic", Name: "Attic"})).To(Succeed())

			cfgs, err := store.ListSceneConfigs(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(cfgs).To(HaveLen(2))
			Expect(cfgs[0].ID).To(Equal("attic"))
			Expect(cfgs[1].ID).To(Equal("cave"))
		})

		It("treats a room that is not set up as an empty room", func() {
			room, err := scenestore.NewConfigStore(sceneStorage, roomStorage, "attic").CurrentRoom(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(room.ID).To(Equal("attic"))
			Expect(room.Devices).To(BeEmpty())
		})
	})
})
