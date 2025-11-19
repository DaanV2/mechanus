package main

import (
	"bytes"
	"strings"
	"text/template"
)

type Settings struct {
	Name        string
	Description string
	Depth       int

	Fields []Field
}

type Field struct {
	Name        string
	Description string
	Default     string
	Env         string
	Type        string
}

const settings_section_template = `{{ $header := repeat "#" .Depth }}
{{$header}} {{.Name}}

{{- if .Description}}
{{.Description}}

{{- end}}
{{if gt (len .Fields) 0}}
| Name | Type | Description | Default | Env |
|------|------|-------------|---------|-----|
{{range .Fields}}| {{.Name}} | {{.Type}} | {{.Description}} | {{.Default}} | {{.Env}} |
{{end}}{{end}}`


func (s *Settings) render() ([]byte, error) {
	v := template.New("settings").Funcs(template.FuncMap{"repeat": strings.Repeat})

	templ, err := v.Parse(settings_section_template)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	err = templ.Execute(buf, s)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
