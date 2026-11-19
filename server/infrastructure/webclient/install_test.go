package webclient_test

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing/fstest"

	"github.com/DaanV2/mechanus/server/infrastructure/webclient"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func clientFS(index string) fstest.MapFS {
	return fstest.MapFS{
		"index.html":       {Data: []byte(index)},
		"_app/app.js":      {Data: []byte("console.log('app')")},
		"licenses/go.json": {Data: []byte(`[{"license":"MIT"}]`)},
	}
}

func readServed(fsys http.FileSystem, name string) string {
	f, err := fsys.Open(name)
	Expect(err).ToNot(HaveOccurred())
	defer func() { _ = f.Close() }()
	data, err := io.ReadAll(f)
	Expect(err).ToNot(HaveOccurred())

	return string(data)
}

var _ = Describe("Install", func() {
	var (
		folder string
	)

	BeforeEach(func() {
		folder = filepath.Join(GinkgoT().TempDir(), "web")
	})

	It("creates and fills a missing folder", func(ctx SpecContext) {
		served := webclient.Install(ctx, folder, clientFS("v1"))

		Expect(filepath.Join(folder, "index.html")).To(BeAnExistingFile())
		Expect(filepath.Join(folder, "_app", "app.js")).To(BeAnExistingFile())
		Expect(filepath.Join(folder, webclient.MarkerFile)).To(BeAnExistingFile())
		Expect(readServed(served, "/index.html")).To(Equal("v1"))
	})

	It("fills an empty folder", func(ctx SpecContext) {
		Expect(os.MkdirAll(folder, 0o750)).To(Succeed())

		served := webclient.Install(ctx, folder, clientFS("v1"))

		Expect(readServed(served, "/index.html")).To(Equal("v1"))
		Expect(filepath.Join(folder, webclient.MarkerFile)).To(BeAnExistingFile())
	})

	It("replaces a folder filled by another version", func(ctx SpecContext) {
		webclient.Install(ctx, folder, clientFS("v1"))
		Expect(os.WriteFile(filepath.Join(folder, "stale.js"), []byte("old"), 0o600)).To(Succeed())

		served := webclient.Install(ctx, folder, clientFS("v2"))

		Expect(readServed(served, "/index.html")).To(Equal("v2"))
		Expect(filepath.Join(folder, "stale.js")).ToNot(BeAnExistingFile())
	})

	It("redoes an interrupted install", func(ctx SpecContext) {
		Expect(os.MkdirAll(folder, 0o750)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(folder, webclient.MarkerFile), []byte("installing\n"), 0o600)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(folder, "index.html"), []byte("partial"), 0o600)).To(Succeed())

		served := webclient.Install(ctx, folder, clientFS("v1"))

		Expect(readServed(served, "/index.html")).To(Equal("v1"))
	})

	It("keeps a folder filled by the same version", func(ctx SpecContext) {
		webclient.Install(ctx, folder, clientFS("v1"))
		Expect(os.WriteFile(filepath.Join(folder, "extra.txt"), []byte("kept"), 0o600)).To(Succeed())

		webclient.Install(ctx, folder, clientFS("v1"))

		Expect(filepath.Join(folder, "extra.txt")).To(BeAnExistingFile())
	})

	It("leaves a folder without a marker untouched", func(ctx SpecContext) {
		Expect(os.MkdirAll(folder, 0o750)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(folder, "index.html"), []byte("custom"), 0o600)).To(Succeed())

		served := webclient.Install(ctx, folder, clientFS("v1"))

		Expect(readServed(served, "/index.html")).To(Equal("custom"))
		Expect(filepath.Join(folder, webclient.MarkerFile)).ToNot(BeAnExistingFile())
	})

	It("serves the embedded client when the folder can't be created", func(ctx SpecContext) {
		file := filepath.Join(GinkgoT().TempDir(), "file")
		Expect(os.WriteFile(file, []byte("not a folder"), 0o600)).To(Succeed())

		served := webclient.Install(ctx, filepath.Join(file, "web"), clientFS("v1"))

		Expect(readServed(served, "/index.html")).To(Equal("v1"))
	})

	It("serves the folder as is without an embedded client", func(ctx SpecContext) {
		served := webclient.Install(ctx, folder, nil)

		Expect(folder).ToNot(BeADirectory())
		_, err := served.Open("/index.html")
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("Hash", func() {
	It("changes with the content", func() {
		a, err := webclient.Hash(clientFS("v1"))
		Expect(err).ToNot(HaveOccurred())
		b, err := webclient.Hash(clientFS("v2"))
		Expect(err).ToNot(HaveOccurred())

		Expect(a).ToNot(Equal(b))
	})
})
