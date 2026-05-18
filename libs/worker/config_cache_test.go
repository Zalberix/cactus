package worker

import "testing"

func TestConfigCacheKeyIncludesScopeRevisionAndHash(t *testing.T) {
	ref := ConfigRef{OrganizationID: 12, WorkTypeID: 3, RevisionID: 73, ConfigHash: "sha256:abc"}
	if got := configCacheKey(ref); got != "12:3:73:sha256:abc" {
		t.Fatalf("configCacheKey() = %q", got)
	}
}

func TestConfigCacheSeparatesHashChanges(t *testing.T) {
	cache := newConfigCache()
	ref := ConfigRef{OrganizationID: 12, WorkTypeID: 3, RevisionID: 73, ConfigHash: "sha256:old"}
	cache.set(ref, map[string]any{"host": "old"})

	if _, ok := cache.get(ConfigRef{OrganizationID: 12, WorkTypeID: 3, RevisionID: 73, ConfigHash: "sha256:new"}); ok {
		t.Fatal("expected hash change to miss cache")
	}
}
