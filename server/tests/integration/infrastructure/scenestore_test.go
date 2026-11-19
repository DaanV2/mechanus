package infrastructure_test

import (
	"context"

	"github.com/DaanV2/mechanus/server/application"
	"github.com/DaanV2/mechanus/server/engine/scenes"
	"github.com/DaanV2/mechanus/server/infrastructure/scenestore"
	util_test "github.com/DaanV2/mechanus/server/tests/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ application.SceneStateStore = (*scenestore.StateStore)(nil)

var _ = Describe("scenestore StateStore", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	It("round trips scene state through the database", func() {
		store, err := scenestore.NewDBStateStore(util_test.CreateIntegrationDatabase(ctx))
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
