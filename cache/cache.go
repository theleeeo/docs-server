package cache

import "fmt"

type Cache struct {
	cache map[string][]byte
}

func New() *Cache {
	return &Cache{
		cache: make(map[string][]byte),
	}
}

func getCacheKey(version, file string) string {
	return fmt.Sprint(version, "/", file)
}

func (p *Cache) loadFromCache(version, file string) ([]byte, error) {
	key := getCacheKey(version, file)
	data, ok := p.cache[key]
	if !ok {
		return nil, nil
	}
	return data, nil
}

func (p *Cache) saveToCache(version, file string, data []byte) {
	key := getCacheKey(version, file)
	p.cache[key] = data
}

func (p *Cache) Get(version, file string) ([]byte, error) {
	return p.loadFromCache(version, file)
}

func (p *Cache) Set(version, file string, data []byte) error {
	p.saveToCache(version, file, data)
	return nil
}
