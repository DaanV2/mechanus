package logging

import (
	"context"

	"github.com/charmbracelet/log"
)

type Enriched struct {
	prefix string
	values []any
}

func (e Enriched) WithPrefix(prefix string) Enriched {
	e.prefix = prefix
	return e
}

func (e Enriched) With(keyvalues ...interface{}) Enriched {
	e.values = append(e.values, keyvalues...)
	return e
}

func (p Enriched) From(ctx context.Context) *log.Logger {
	l := From(ctx)
	if len(p.prefix) > 0 {
		l = l.WithPrefix(p.prefix)
	}
	if len(p.values) > 0 {
		l = l.With(p.values...)
	}

	return l
}
