package main

import (
	"bytes"
	"fmt"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"github.com/DaanV2/mechanus/server/cmd"
	"github.com/DaanV2/mechanus/server/internal/setup"
	"github.com/DaanV2/mechanus/server/pkg/config"
	xos "github.com/DaanV2/mechanus/server/pkg/extensions/os"
	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	output_dir = filepath.Join(".", "docs", "cmd")
)

type markdownBuilder struct {
	sections []*Settings
}

func (b *markdownBuilder) add(s ...*Settings) {
	b.sections = append(b.sections, s...)
}

func (b *markdownBuilder) render() ([]byte, error) {
	data := new(bytes.Buffer)

	for _, s := range b.sections {
		d, err := s.render()
		if err != nil {
			return nil, err
		}
		_, err = data.Write(d)
		if err != nil {
			return nil, err
		}
	}

	return data.Bytes(), nil
}

func main() {
	setup.Viper()
	setup.Logger()

	_ = cmd.RootCommand()

	builder := &markdownBuilder{}
	settings := viper.AllSettings()

	markdownStruct(builder, "settings", "", settings, 1)

	data, err := builder.render()
	if err != nil {
		log.Fatal("troubling rendering settings", "error", err)
	}
	err = xos.WriteFile(filepath.Join(".", "docs", "settings.md"), data)
	if err != nil {
		log.Fatal("troubling rendering settings", "error", err)
	}
}

func markdownStruct(builder *markdownBuilder, key, configKey string, value map[string]any, depth int) {
	s := &Settings{
		Name:        cases.Title(language.BritishEnglish).String(key),
		Description: "",
		Depth:       depth,
		Fields:      make([]Field, 0),
	}
	builder.add(s)

	l := make([]Field, 0)

	for k, v := range value {
		switch item := v.(type) {
		case map[string]any:
			ck := strings.Trim(configKey+"."+k, " .")

			markdownStruct(builder, k, ck, item, depth+1)
			l = append(l, Field{
				Name:        k,
				Description: fmt.Sprintf("see: [%s](#%s)", k, k),
				Type:        "object",
			})

		default:
			ck := strings.Trim(configKey+"."+k, " .")
			s.Fields = append(s.Fields, createField(k, ck, item))
		}
	}

	slices.SortFunc(s.Fields, func(a Field, b Field) int {
		return strings.Compare(a.Name, b.Name)
	})
	slices.SortFunc(l, func(a Field, b Field) int {
		return strings.Compare(a.Name, b.Name)
	})

	// Add object to the end
	s.Fields = append(s.Fields, l...)
}

func createField(key, configKey string, value any) Field {

	result := Field{
		Name:        key,
		Description: "",
		Default:     fmt.Sprintf("%v", value),
		Env:         "",
		Type:        reflect.TypeOf(value).Name(),
	}

	f := findFlags(configKey)
	if f == nil {
		log.Error("cannot find flag", "config", configKey)
		return result
	}

	result.Env = config.EnvironmentNamer().Replace(strings.ToUpper(f.Name()))
	result.Type = strings.ToLower(f.Type())
	result.Description = f.Description()

	return result
}

func findFlags(name string) config.BaseFlag {
	confs := config.AllConfigs()

	for _, c := range confs {
		for _, f := range c.All() {
			if strings.EqualFold(f.Name(), name) {
				return f
			}
		}
	}

	return nil
}
