// Command licenses writes the licences of every Go module compiled into the mechanus binary to a JSON file.
// The client's About page reads that file (served at /licenses/go.json) to list third-party licences.
//
// Usage (from server/): go run ./scripts/licenses -o ../client/build/licenses/go.json
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/charmbracelet/log"
)

// Entry is one third-party dependency and its licence. The client's npm licence list uses the same shape.
type Entry struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	License string `json:"license"`
	URL     string `json:"url,omitempty"`
	Text    string `json:"text,omitempty"`
}

type goModule struct {
	Path    string
	Version string
	Dir     string
	Main    bool
	Replace *goModule
}

type goPackage struct {
	Module *goModule
}

func main() {
	output := flag.String("o", "licenses-go.json", "file to write the licence list to")
	flag.Parse()

	modules, err := listModules()
	if err != nil {
		log.Fatal("listing Go modules", "error", err)
	}

	entries := make([]Entry, 0, len(modules))
	for _, mod := range modules {
		entries = append(entries, entryFor(mod))
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		log.Fatal("encoding licences", "error", err)
	}

	err = os.MkdirAll(filepath.Dir(*output), 0o750)
	if err == nil {
		//nolint:gosec // G306: the licence list is a public web asset, served to every client
		err = os.WriteFile(*output, data, 0o644)
	}
	if err != nil {
		log.Fatal("writing licences", "file", *output, "error", err)
	}

	log.Info("wrote Go licences", "file", *output, "modules", len(entries))
}

// listModules returns the modules that provide packages linked into the server (test and tool dependencies excluded).
func listModules() ([]goModule, error) {
	cmd := exec.CommandContext(context.Background(), "go", "list", "-deps", "-json=Module", "./...")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list: %w", err)
	}

	seen := map[string]goModule{}
	dec := json.NewDecoder(bytes.NewReader(out))
	for {
		var pkg goPackage
		err := dec.Decode(&pkg)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("decoding go list output: %w", err)
		}

		if pkg.Module == nil || pkg.Module.Main {
			continue // standard library or mechanus itself
		}
		mod := *pkg.Module
		if mod.Replace != nil {
			mod.Version = mod.Replace.Version
			mod.Dir = mod.Replace.Dir
		}
		seen[mod.Path] = mod
	}

	modules := make([]goModule, 0, len(seen))
	for _, mod := range seen {
		modules = append(modules, mod)
	}
	slices.SortFunc(modules, func(a, b goModule) int { return strings.Compare(a.Path, b.Path) })

	return modules, nil
}

func entryFor(mod goModule) Entry {
	entry := Entry{
		Name:    mod.Path,
		Version: mod.Version,
		License: "Unknown",
		URL:     "https://pkg.go.dev/" + mod.Path,
	}

	texts := licenseTexts(mod.Dir)
	if len(texts) == 0 {
		log.Warn("no licence file found", "module", mod.Path)

		return entry
	}

	entry.License = Detect(texts[0])
	entry.Text = strings.Join(texts, "\n\n")

	return entry
}

// licenseTexts reads the licence and notice files in the root of a module.
func licenseTexts(dir string) []string {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil
	}
	defer func() { _ = root.Close() }()

	files, err := fs.ReadDir(root.FS(), ".")
	if err != nil {
		return nil
	}

	var licenses, notices []string
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := strings.ToUpper(f.Name())
		isLicense := strings.HasPrefix(name, "LICENSE") || strings.HasPrefix(name, "LICENCE") || strings.HasPrefix(name, "COPYING")
		isNotice := strings.HasPrefix(name, "NOTICE")
		if !isLicense && !isNotice {
			continue
		}

		data, err := root.ReadFile(f.Name())
		if err != nil {
			continue
		}
		text := strings.TrimSpace(string(data))
		if isLicense {
			licenses = append(licenses, text)
		} else {
			notices = append(notices, text)
		}
	}

	return append(licenses, notices...)
}

// Detect guesses the SPDX identifier of a licence text from phrases that only appear in that licence.
func Detect(text string) string {
	t := strings.Join(strings.Fields(text), " ")
	contains := func(s string) bool { return strings.Contains(t, s) }

	switch {
	case contains("Apache License") && contains("Version 2.0"):
		return "Apache-2.0"
	case contains("Eclipse Public License") && contains("Eclipse Distribution License"):
		return "EPL-2.0 OR BSD-3-Clause"
	case contains("Eclipse Public License"):
		return "EPL-2.0"
	case contains("Mozilla Public License") && contains("2.0"):
		return "MPL-2.0"
	case contains("Permission is hereby granted, free of charge"):
		return "MIT"
	case contains("Redistribution and use in source and binary forms"):
		if contains("Neither the name") || contains("names of its contributors") {
			return "BSD-3-Clause"
		}

		return "BSD-2-Clause"
	case contains("Permission to use, copy, modify, and/or distribute this software for any purpose"),
		contains("Permission to use, copy, modify, and distribute this software for any purpose"):
		return "ISC"
	case contains("This is free and unencumbered software released into the public domain"):
		return "Unlicense"
	case contains("GNU LESSER GENERAL PUBLIC LICENSE"):
		return "LGPL"
	case contains("GNU GENERAL PUBLIC LICENSE"):
		return "GPL"
	default:
		return "Unknown"
	}
}
