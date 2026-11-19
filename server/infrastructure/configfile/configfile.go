// Package configfile decodes authored configuration files (scenes, rooms): YAML or JSON chosen by file
// extension, with a required `version` field and a chain of migrations that upgrade older files in
// memory. Files are never rewritten.
package configfile

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// VersionField is the name of the field every config file carries its format version in.
const VersionField = "version"

var (
	// ErrUnsupportedFormat is returned for a file extension that is not a config file format.
	ErrUnsupportedFormat = errors.New("unsupported config file format")
	// ErrMissingVersion is returned for a file without a valid `version` field.
	ErrMissingVersion = errors.New("missing or invalid version")
	// ErrUnsupportedVersion is returned for a file with a version newer than this server supports.
	ErrUnsupportedVersion = errors.New("unsupported version")
)

// extensions are the supported file extensions, lower case.
var extensions = []string{".yaml", ".yml", ".json"}

// Extensions returns the supported file extensions.
func Extensions() []string {
	return slices.Clone(extensions)
}

// IsConfigFile reports whether the file name has a supported config file extension.
func IsConfigFile(name string) bool {
	return slices.Contains(extensions, strings.ToLower(filepath.Ext(name)))
}

// ID returns the file name without its extension: the ID of the scene or room the file defines.
func ID(name string) string {
	base := filepath.Base(name)

	return strings.TrimSuffix(base, filepath.Ext(base))
}

// Migration upgrades a decoded document from one version to the next. It may modify doc in place.
type Migration func(doc map[string]any) error

// Migrations is the ordered chain of migrations of one kind of config file. Versions start at 1:
// Migrations[0] upgrades version 1 to 2, Migrations[1] version 2 to 3, and so on. The current version
// is therefore len(Migrations)+1.
type Migrations []Migration

// Current returns the current (newest) version.
func (m Migrations) Current() int {
	return len(m) + 1
}

// Decode parses data as the format of the file name's extension, migrates it to the current version
// and decodes the result into out using its JSON field names, so YAML and JSON files share one schema.
func Decode(name string, data []byte, migrations Migrations, out any) error {
	doc, err := parse(name, data)
	if err != nil {
		return err
	}

	version, err := versionOf(doc)
	if err != nil {
		return err
	}
	if version > migrations.Current() {
		return fmt.Errorf("%w: %d, this server supports up to %d", ErrUnsupportedVersion, version, migrations.Current())
	}
	for v := version; v < migrations.Current(); v++ {
		if err := migrations[v-1](doc); err != nil {
			return fmt.Errorf("migrating from version %d to %d: %w", v, v+1, err)
		}
	}
	delete(doc, VersionField)

	normalized, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("normalizing document: %w", err)
	}
	if err := json.Unmarshal(normalized, out); err != nil {
		return fmt.Errorf("decoding: %w", err)
	}

	return nil
}

// Encode returns v as a file of the format of the file name's extension, with the given version.
func Encode(name string, version int, v any) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("%T is not an object: %w", v, err)
	}
	doc[VersionField] = version

	switch strings.ToLower(filepath.Ext(name)) {
	case ".json":
		return json.MarshalIndent(doc, "", "  ")
	case ".yaml", ".yml":
		var buf bytes.Buffer
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(2)
		if err := enc.Encode(doc); err != nil {
			return nil, err
		}

		return buf.Bytes(), enc.Close()
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnsupportedFormat, filepath.Ext(name))
	}
}

func parse(name string, data []byte) (map[string]any, error) {
	var doc map[string]any
	switch strings.ToLower(filepath.Ext(name)) {
	case ".json":
		if err := json.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("parsing JSON: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("parsing YAML: %w", err)
		}
	default:
		return nil, fmt.Errorf("%w: %q, use %s", ErrUnsupportedFormat, filepath.Ext(name), strings.Join(extensions, ", "))
	}
	if doc == nil {
		return nil, fmt.Errorf("%w: empty file", ErrMissingVersion)
	}

	return doc, nil
}

// versionOf reads the version field. JSON numbers decode as float64, YAML numbers as int.
func versionOf(doc map[string]any) (int, error) {
	raw, ok := doc[VersionField]
	if !ok {
		return 0, ErrMissingVersion
	}

	var version int
	switch v := raw.(type) {
	case int:
		version = v
	case float64:
		if v != math.Trunc(v) || v > math.MaxInt32 {
			return 0, fmt.Errorf("%w: %v", ErrMissingVersion, v)
		}
		version = int(v)
	default:
		return 0, fmt.Errorf("%w: %v", ErrMissingVersion, raw)
	}
	if version < 1 {
		return 0, fmt.Errorf("%w: %d", ErrMissingVersion, version)
	}

	return version, nil
}
