package assets_test

import (
	"net/http"
	"net/http/httptest"
	"testing/fstest"

	"github.com/DaanV2/mechanus/server/infrastructure/assets"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Store", func() {
	var (
		files fstest.MapFS
		store *assets.Store
	)

	BeforeEach(func() {
		files = fstest.MapFS{
			"scenes/cave.dat":                      {Data: []byte(`{"id":"cave"}`)},
			"scenes/cave/maps/cave.png":            {Data: []byte("cave map")},
			"shared/monsters/goblins/goblin.png":   {Data: []byte("goblin")},
			"keys/secret.dat":                      {Data: []byte("secret")},
			"scenes/cave/maps/folder/.placeholder": {Data: []byte("")},
		}
		store = assets.NewStore(files)
	})

	Describe("Resolve", func() {
		It("resolves a path relative to the scene folder", func() {
			r, err := store.Resolve("cave", "maps/cave.png")
			Expect(err).ToNot(HaveOccurred())
			Expect(r.URL).To(MatchRegexp(`^/assets/scenes/cave/maps/cave\.png\?v=[0-9a-f]{16}$`))
			Expect(r.MediaType).To(Equal("image/png"))
		})

		It("resolves shared assets", func() {
			r, err := store.Resolve("cave", "shared:/monsters/goblins/goblin.png")
			Expect(err).ToNot(HaveOccurred())
			Expect(r.URL).To(HavePrefix("/assets/shared/monsters/goblins/goblin.png?v="))
		})

		It("passes absolute URLs through", func() {
			r, err := store.Resolve("cave", "https://example.com/map.webp")
			Expect(err).ToNot(HaveOccurred())
			Expect(r.URL).To(Equal("https://example.com/map.webp"))
		})

		It("changes the hash when the file changes", func() {
			before, err := store.Resolve("cave", "maps/cave.png")
			Expect(err).ToNot(HaveOccurred())

			files["scenes/cave/maps/cave.png"] = &fstest.MapFile{Data: []byte("a bigger cave map")}
			after, err := store.Resolve("cave", "maps/cave.png")
			Expect(err).ToNot(HaveOccurred())
			Expect(after.URL).ToNot(Equal(before.URL))
		})

		DescribeTable("rejects references outside the asset folders",
			func(sceneID, ref string) {
				_, err := store.Resolve(sceneID, ref)
				Expect(err).To(HaveOccurred())
			},
			Entry("parent of the scene folder", "cave", "../cave.dat"),
			Entry("outside the config folder", "cave", "../../keys/secret.dat"),
			Entry("parent of the shared folder", "cave", "shared:/../keys/secret.dat"),
			Entry("unsafe scene id", "../keys", "secret.dat"),
			Entry("missing file", "cave", "maps/missing.png"),
			Entry("folder", "cave", "maps/folder"),
			Entry("empty reference", "cave", ""),
		)
	})

	Describe("ServeHTTP", func() {
		get := func(target string) *httptest.ResponseRecorder {
			rec := httptest.NewRecorder()
			store.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, http.NoBody))

			return rec
		}

		It("serves scene assets, cacheable when versioned", func() {
			rec := get("/assets/scenes/cave/maps/cave.png?v=abc")
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Body.String()).To(Equal("cave map"))
			Expect(rec.Header().Get("Cache-Control")).To(ContainSubstring("immutable"))
		})

		It("serves shared assets", func() {
			rec := get("/assets/shared/monsters/goblins/goblin.png")
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Header().Get("Cache-Control")).To(Equal("no-cache"))
		})

		DescribeTable("does not serve anything outside the asset folders",
			func(target string) {
				Expect(get(target).Code).To(Equal(http.StatusNotFound))
			},
			Entry("scene config", "/assets/scenes/cave.dat"),
			Entry("other config folders", "/assets/keys/secret.dat"),
			Entry("folder", "/assets/scenes/cave/maps"),
			Entry("missing file", "/assets/shared/missing.png"),
		)
	})
})
