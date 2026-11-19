package configfile_test

import (
	"github.com/DaanV2/mechanus/server/infrastructure/configfile"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type doc struct {
	ID       string   `json:"id"`
	Name     string   `json:"display_name"`
	Tags     []string `json:"tags"`
	Renamed  string   `json:"renamed,omitempty"`
	Optional *int     `json:"optional,omitempty"`
}

var _ = Describe("configfile", func() {
	const yamlFile = `
version: 1
id: cave
display_name: The cave
tags: [dark, wet]
`
	const jsonFile = `{"version": 1, "id": "cave", "display_name": "The cave", "tags": ["dark", "wet"]}`

	It("decodes YAML and JSON into the same result", func() {
		var fromYAML, fromJSON doc
		Expect(configfile.Decode("cave.yaml", []byte(yamlFile), nil, &fromYAML)).To(Succeed())
		Expect(configfile.Decode("cave.json", []byte(jsonFile), nil, &fromJSON)).To(Succeed())

		Expect(fromYAML).To(Equal(doc{ID: "cave", Name: "The cave", Tags: []string{"dark", "wet"}}))
		Expect(fromJSON).To(Equal(fromYAML))
	})

	It("accepts the .yml extension and is case insensitive", func() {
		var d doc
		Expect(configfile.Decode("CAVE.YML", []byte(yamlFile), nil, &d)).To(Succeed())
		Expect(d.ID).To(Equal("cave"))
	})

	It("rejects other extensions", func() {
		var d doc
		Expect(configfile.Decode("cave.dat", []byte(jsonFile), nil, &d)).To(MatchError(configfile.ErrUnsupportedFormat))
	})

	It("reports syntax errors", func() {
		var d doc
		Expect(configfile.Decode("cave.yaml", []byte("id: [unclosed"), nil, &d)).ToNot(Succeed())
		Expect(configfile.Decode("cave.json", []byte(`{"id":`), nil, &d)).ToNot(Succeed())
	})

	DescribeTable("rejects files without a valid version",
		func(name, content string) {
			var d doc
			Expect(configfile.Decode(name, []byte(content), nil, &d)).To(MatchError(configfile.ErrMissingVersion))
		},
		Entry("missing", "cave.yaml", "id: cave"),
		Entry("zero", "cave.yaml", "version: 0\nid: cave"),
		Entry("text", "cave.yaml", "version: one\nid: cave"),
		Entry("fraction", "cave.json", `{"version": 1.5}`),
		Entry("empty file", "cave.yaml", ""),
	)

	Describe("migrations", func() {
		// Version 1 called the field "title", version 2 "display_name"; version 3 added "renamed".
		migrations := configfile.Migrations{
			func(d map[string]any) error {
				d["display_name"] = d["title"]
				delete(d, "title")

				return nil
			},
			func(d map[string]any) error {
				d["renamed"] = "yes"

				return nil
			},
		}

		It("knows the current version", func() {
			Expect(migrations.Current()).To(Equal(3))
			Expect(configfile.Migrations(nil).Current()).To(Equal(1))
		})

		It("migrates an older file in order", func() {
			var d doc
			Expect(configfile.Decode("cave.yaml", []byte("version: 1\nid: cave\ntitle: Old cave"), migrations, &d)).To(Succeed())
			Expect(d).To(Equal(doc{ID: "cave", Name: "Old cave", Renamed: "yes"}))
		})

		It("only runs the migrations a file needs", func() {
			var d doc
			Expect(configfile.Decode("cave.json", []byte(`{"version": 2, "id": "cave", "display_name": "Cave"}`), migrations, &d)).To(Succeed())
			Expect(d).To(Equal(doc{ID: "cave", Name: "Cave", Renamed: "yes"}))
		})

		It("loads a current file unchanged", func() {
			var d doc
			Expect(configfile.Decode("cave.yaml", []byte("version: 3\nid: cave\nrenamed: no"), migrations, &d)).To(Succeed())
			Expect(d.Renamed).To(Equal("no"))
		})

		It("rejects a newer file", func() {
			var d doc
			Expect(configfile.Decode("cave.yaml", []byte("version: 4\nid: cave"), migrations, &d)).To(MatchError(configfile.ErrUnsupportedVersion))
		})
	})

	DescribeTable("round trips through Encode",
		func(name string) {
			one := 1
			in := doc{ID: "cave", Name: "Cave", Tags: []string{"a"}, Optional: &one}
			data, err := configfile.Encode(name, 1, in)
			Expect(err).ToNot(HaveOccurred())

			var out doc
			Expect(configfile.Decode(name, data, nil, &out)).To(Succeed())
			Expect(out).To(Equal(in))
		},
		Entry("YAML", "cave.yaml"),
		Entry("JSON", "cave.json"),
	)

	It("derives IDs and recognises config files from names", func() {
		Expect(configfile.ID("scenes/cave.yaml")).To(Equal("cave"))
		Expect(configfile.IsConfigFile("cave.json")).To(BeTrue())
		Expect(configfile.IsConfigFile("cave.dat")).To(BeFalse())
	})
})
