package ports

import (
	"context"
	"time"

	"zpwoot/internal/domain/media"
)

// MediaRepository defines the interface for media storage operations
type MediaRepository interface {
	// SaveCachedMedia saves a cached media item to storage
	SaveCachedMedia(ctx context.Context, item *media.CachedMediaItem) error

	// GetCachedMedia retrieves a cached media item by session and message ID
	GetCachedMedia(ctx context.Context, sessionID, messageID string) (*media.CachedMediaItem, error)

	// UpdateCachedMedia updates an existing cached media item
	UpdateCachedMedia(ctx context.Context, item *media.CachedMediaItem) error

	// DeleteCachedMedia removes a cached media item from storage
	DeleteCachedMedia(ctx context.Context, sessionID, messageID string) error

	// ListCachedMedia lists cached media items with pagination and filtering
	ListCachedMedia(ctx context.Context, req *media.ListCachedMediaRequest) (*media.ListCachedMediaResponse, error)

	// ClearExpiredCache removes cached media items older than the specified duration
	ClearExpiredCache(ctx context.Context, sessionID string, olderThan time.Duration) (*media.ClearCacheResponse, error)

	// GetMediaStats returns statistics about cached media for a session
	GetMediaStats(ctx context.Context, sessionID string) (*media.MediaStats, error)

	// CleanupOrphanedFiles removes files that exist on disk but not in database
	CleanupOrphanedFiles(ctx context.Context, sessionID string) error

	// GetCacheSize returns the total size of cached media for a session
	GetCacheSize(ctx context.Context, sessionID string) (int64, error)

	// GetCacheCount returns the total number of cached media items for a session
	GetCacheCount(ctx context.Context, sessionID string) (int, error)

	// GetExpiredCacheItems returns cached media items that have expired
	GetExpiredCacheItems(ctx context.Context, sessionID string) ([]*media.CachedMediaItem, error)

	// BulkDeleteCachedMedia removes multiple cached media items
	BulkDeleteCachedMedia(ctx context.Context, sessionID string, messageIDs []string) error

	// GetCachedMediaByType returns cached media items filtered by media type
	GetCachedMediaByType(ctx context.Context, sessionID string, mediaType string) ([]*media.CachedMediaItem, error)

	// UpdateLastAccess updates the last access time for a cached media item
	UpdateLastAccess(ctx context.Context, sessionID, messageID string) error

	// GetLeastRecentlyUsed returns the least recently used cached media items
	GetLeastRecentlyUsed(ctx context.Context, sessionID string, limit int) ([]*media.CachedMediaItem, error)

	// GetCacheSizeByType returns the total size of cached media by type for a session
	GetCacheSizeByType(ctx context.Context, sessionID string) (map[string]int64, error)

	// GetCacheCountByType returns the count of cached media by type for a session
	GetCacheCountByType(ctx context.Context, sessionID string) (map[string]int, error)
}
