// shared/common/tenantdb/cache_test.go
package tenantdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache_GetTenantIDByDomain(t *testing.T) {
	t.Run("returns 0 false for nil cache", func(t *testing.T) {
		var cache *Cache
		tenantID, found := cache.GetTenantIDByDomain("example.com")
		assert.False(t, found)
		assert.Equal(t, 0, tenantID)
	})

	t.Run("returns 0 false for cache with nil provider", func(t *testing.T) {
		cache := &Cache{provider: nil}
		tenantID, found := cache.GetTenantIDByDomain("example.com")
		assert.False(t, found)
		assert.Equal(t, 0, tenantID)
	})
}
