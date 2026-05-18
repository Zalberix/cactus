package worker

import (
	"fmt"
	"sync"
)

type configCache struct {
	mu   sync.RWMutex
	data map[string]map[string]any
}

func newConfigCache() *configCache {
	return &configCache{data: map[string]map[string]any{}}
}

func configCacheKey(ref ConfigRef) string {
	return fmt.Sprintf("%d:%d:%d:%s", ref.OrganizationID, ref.WorkTypeID, ref.RevisionID, ref.ConfigHash)
}

func (c *configCache) get(ref ConfigRef) (map[string]any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.data[configCacheKey(ref)]
	return v, ok
}

func (c *configCache) set(ref ConfigRef, settings map[string]any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[configCacheKey(ref)] = settings
}

func (c *configCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = map[string]map[string]any{}
}
