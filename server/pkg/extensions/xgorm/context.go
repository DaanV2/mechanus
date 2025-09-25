package xgorm

import "context"

type prefixContext struct{}
type silentContext struct{}

func WithPrefix(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, prefixContext{}, name)
}

func WithSilent(ctx context.Context, silent bool) context.Context {
	return context.WithValue(ctx, silentContext{}, silent)
}

func GetPrefix(ctx context.Context) (string, bool) {
	value := ctx.Value(prefixContext{})
	if value == nil {
		return "", false
	}

	str, ok := value.(string)

	return str, ok
}

func IsSilent(ctx context.Context) bool {
	value := ctx.Value(silentContext{})
	if value == nil {
		return false
	}

	b, _ := value.(bool)

	return b
}
