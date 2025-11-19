package xgorm

import "context"

type prefixContext struct{}
type silentContext struct{}

// WithPrefix returns a new context with the given prefix name attached.
func WithPrefix(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, prefixContext{}, name)
}

// WithSilent returns a new context with the silent flag set to the given value.
func WithSilent(ctx context.Context, silent bool) context.Context {
	return context.WithValue(ctx, silentContext{}, silent)
}

// GetPrefix retrieves the prefix value from the context. Returns the prefix and true if found, empty string and false otherwise.
func GetPrefix(ctx context.Context) (string, bool) {
	value := ctx.Value(prefixContext{})
	if value == nil {
		return "", false
	}

	str, ok := value.(string)

	return str, ok
}

// IsSilent returns true if the silent flag is set to true in the context.
func IsSilent(ctx context.Context) bool {
	value := ctx.Value(silentContext{})
	if value == nil {
		return false
	}

	b, _ := value.(bool)

	return b
}
