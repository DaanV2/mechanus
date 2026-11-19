package infrastructure_test

import (
	"context"
	"os"
	"path/filepath"

	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/devices"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/infrastructure/configfile"
	"github.com/DaanV2/mechanus/server/infrastructure/scenestore"
	util_test "github.com/DaanV2/mechanus/server/tests/component-test/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	_ application.SceneConfigStore = (*scenestore.ConfigStore)(nil)
	_ application.RoomConfigStore  = (*scenestore.ConfigStore)(nil)
	_ application.SceneStateStore  = (*scenestore.StateStore)(nil)
)

var _ = Describe("scenestore", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	Describe("StateStore", func() {
		It("round trips scene state through the database", func() {
			store, err := scenestore.NewDBStateStore(util_test.CreateDatabase(ctx))
			Expect(err).ToNot(HaveOccurred())

			_, err = store.LoadSceneState(ctx, "cave")
			Expect(err).To(MatchError(scenes.ErrStateNotFound))

			state := &scenes.State{
				SceneID:   "cave",
				VariantID: "collapse",
				Entities:  map[string]scenes.EntityState{"troll": {Hidden: false, Removed: true}},
				Spawned:   []scenes.Entity{{ID: "bat", Kind: scenes.EntityEnemy, Name: "Bat"}},
				Overrides: scenes.Instructions{Lights: map[string]scenes.LightInstruction{"ambient": {On: false}}},
			}
			Expect(store.SaveSceneState(ctx, state)).To(Succeed())

			loaded, err := store.LoadSceneState(ctx, "cave")
			Expect(err).ToNot(HaveOccurred())
			Expect(loaded).To(Equal(state))
		})
	})

	Describe("ConfigStore", func() {
		var (
			store *scenestore.ConfigStore
			dir   string
		)

		BeforeEach(func() {
			store, dir = util_test.CreateConfigStore(ctx, devices.NewRoom("living-room", "Living room"),
				&scenes.Config{ID: "cave", Name: "Cave"},
			)
		})

		sceneName := func(id string) string {
			GinkgoHelper()
			cfg, err := store.GetSceneConfig(ctx, id)
			Expect(err).ToNot(HaveOccurred())

			return cfg.Name
		}

		It("reads scene configs", func() {
			Expect(sceneName("cave")).To(Equal("Cave"))
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
			util_test.WriteConfigFile(dir, "scenes/attic.json", &scenes.Config{ID: "attic", Name: "Attic"})
			store.Load(ctx)

			cfgs, err := store.ListSceneConfigs(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(cfgs).To(HaveLen(2))
			Expect(cfgs[0].ID).To(Equal("attic"))
			Expect(cfgs[1].ID).To(Equal("cave"))
		})

		It("treats a room that is not set up as an empty room", func() {
			empty := scenestore.NewConfigStore(dir, "attic")
			empty.Load(ctx)
			room, err := empty.CurrentRoom(ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(room.ID).To(Equal("attic"))
			Expect(room.Devices).To(BeEmpty())
		})

		Describe("formats", func() {
			It("loads YAML and JSON scenes the same", func() {
				util_test.WriteRawFile(dir, "scenes/road.yaml", `
version: 1
id: road
name: Road
entities:
  - { id: cart, kind: object, name: Cart }
variants:
  - id: rain
    name: Rain
    tags: { weather: rain }
    instructions:
      screens:
        table: { view: scenery, asset: images/rain.webp }
`)
				util_test.WriteRawFile(dir, "scenes/lane.json", `{
  "version": 1, "id": "lane", "name": "Road",
  "entities": [{"id": "cart", "kind": "object", "name": "Cart"}],
  "variants": [{"id": "rain", "name": "Rain", "tags": {"weather": "rain"},
    "instructions": {"screens": {"table": {"view": "scenery", "asset": "images/rain.webp"}}}}]
}`)
				store.Load(ctx)
				Expect(store.Errors()).To(BeEmpty())

				road, err := store.GetSceneConfig(ctx, "road")
				Expect(err).ToNot(HaveOccurred())
				lane, err := store.GetSceneConfig(ctx, "lane")
				Expect(err).ToNot(HaveOccurred())
				laneAsRoad := *lane
				laneAsRoad.ID = "road"
				Expect(*road).To(Equal(laneAsRoad))
				Expect(road.Variants[0].Instructions.Screens["table"].Asset).To(Equal("images/rain.webp"))
			})

			It("takes the id from the file name when the file has none", func() {
				util_test.WriteRawFile(dir, "scenes/road.yml", "version: 1\nname: Road")
				store.Load(ctx)
				Expect(sceneName("road")).To(Equal("Road"))
			})

			It("reports a file declaring another id", func() {
				util_test.WriteRawFile(dir, "scenes/road.yaml", "version: 1\nid: lane")
				store.Load(ctx)
				Expect(store.Errors()).To(ConsistOf(HaveField("File", "scenes/road.yaml")))
			})

			It("reports legacy .dat files instead of ignoring them", func() {
				util_test.WriteRawFile(dir, "scenes/old.dat", `{"id": "old"}`)
				store.Load(ctx)
				Expect(store.Errors()).To(ConsistOf(SatisfyAll(
					HaveField("File", "scenes/old.dat"),
					HaveField("Message", ContainSubstring(".yaml or .json")),
				)))
			})

			It("reports two files for the same scene", func() {
				util_test.WriteConfigFile(dir, "scenes/cave.json", &scenes.Config{ID: "cave", Name: "Other cave"})
				store.Load(ctx)
				Expect(store.Errors()).To(ConsistOf(HaveField("Message", ContainSubstring("more than one file"))))
				_, err := store.GetSceneConfig(ctx, "cave")
				Expect(err).To(MatchError(scenes.ErrSceneNotFound), "the scene is skipped")
			})

			It("ignores asset folders and hidden files", func() {
				util_test.WriteRawFile(dir, "scenes/cave/maps/cave.map.json", `{}`)
				util_test.WriteRawFile(dir, "scenes/.cave.yaml.swp", "garbage")
				store.Load(ctx)
				Expect(store.Errors()).To(BeEmpty())
			})
		})

		Describe("versions", func() {
			It("rejects a file without a version", func() {
				util_test.WriteRawFile(dir, "scenes/road.yaml", "id: road")
				store.Load(ctx)
				Expect(store.Errors()).To(ConsistOf(HaveField("Message", ContainSubstring("version"))))
			})

			It("rejects a file newer than the server supports", func() {
				util_test.WriteRawFile(dir, "scenes/road.yaml", "version: 99\nid: road")
				store.Load(ctx)
				Expect(store.Errors()).To(ConsistOf(HaveField("Message", ContainSubstring(configfile.ErrUnsupportedVersion.Error()))))
				_, err := store.GetSceneConfig(ctx, "road")
				Expect(err).To(MatchError(scenes.ErrSceneNotFound))
			})

			It("migrates an older file in memory without rewriting it", func() {
				// Pretend version 2 renamed "title" to "name"
				original := scenestore.SceneMigrations
				scenestore.SceneMigrations = configfile.Migrations{func(doc map[string]any) error {
					doc["name"] = doc["title"]
					delete(doc, "title")

					return nil
				}}
				DeferCleanup(func() { scenestore.SceneMigrations = original })

				const content = "version: 1\nid: road\ntitle: Old road"
				util_test.WriteRawFile(dir, "scenes/road.yaml", content)
				util_test.WriteRawFile(dir, "scenes/lane.yaml", "version: 2\nid: lane\nname: Lane")
				store.Load(ctx)

				Expect(sceneName("road")).To(Equal("Old road"))
				Expect(sceneName("lane")).To(Equal("Lane"))
				data, err := os.ReadFile(filepath.Join(dir, "scenes", "road.yaml")) //nolint:gosec // a file this test wrote
				Expect(err).ToNot(HaveOccurred())
				Expect(string(data)).To(Equal(content))
			})
		})

		Describe("invalid files", func() {
			It("skips an invalid file, reports it, and loads it once fixed", func() {
				util_test.WriteRawFile(dir, "scenes/cave.yaml", "version: 1\nname: [broken")
				store.Load(ctx)

				_, err := store.GetSceneConfig(ctx, "cave")
				Expect(err).To(MatchError(scenes.ErrSceneNotFound))
				Expect(store.Errors()).To(ConsistOf(HaveField("File", "scenes/cave.yaml")))

				util_test.WriteConfigFile(dir, "scenes/cave.yaml", &scenes.Config{ID: "cave", Name: "Fixed cave"})
				store.Load(ctx)
				Expect(sceneName("cave")).To(Equal("Fixed cave"))
				Expect(store.Errors()).To(BeEmpty())
			})

			It("forgets a deleted scene on the next load", func() {
				Expect(os.Remove(filepath.Join(dir, "scenes", "cave.yaml"))).To(Succeed())
				store.Load(ctx)

				_, err := store.GetSceneConfig(ctx, "cave")
				Expect(err).To(MatchError(scenes.ErrSceneNotFound))
			})
		})
	})
})

var _ = Describe("example config files", func() {
	It("load without errors", func(ctx SpecContext) {
		store := scenestore.NewConfigStore(filepath.Join("..", "..", "..", "..", "docs", "examples", "battlemap"), scenestore.RoomID)
		store.Load(ctx)
		Expect(store.Errors()).To(BeEmpty())

		tavern, err := store.GetSceneConfig(ctx, "example-tavern")
		Expect(err).ToNot(HaveOccurred())
		Expect(tavern.Entities).To(HaveLen(3))
		_, ok := tavern.Variant("night")
		Expect(ok).To(BeTrue())

		room, err := store.CurrentRoom(ctx)
		Expect(err).ToNot(HaveOccurred())
		Expect(room.Bindings["table"]).To(Equal([]string{"example-tv"}))
	})
})
