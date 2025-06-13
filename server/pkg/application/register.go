package application

func Register[T any](c *ComponentManager, v T) T {
	c.Add(v)

	return v
}
