package loader

import "context"

type Cache interface {
	Push(context.Context, string, []byte) error
	Pull(context.Context, string) ([]byte, error)
}

type mapCache struct {
	data map[string][]byte
}

func NewMapCache() Cache {
	return mapCache{
		data: make(map[string][]byte),
	}
}

func (c mapCache) Push(ctx context.Context, key string, value []byte) error {
	c.data[key] = value
	return nil
}

func (c mapCache) Pull(ctx context.Context, key string) ([]byte, error) {
	value, ok := c.data[key]
	if !ok {
		return nil, nil
	}
	return value, nil
}
