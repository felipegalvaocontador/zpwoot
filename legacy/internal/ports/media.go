package ports

import (
	"context"
	"time"

	"zpwoot/internal/domain/media"
)

type MediaRepository interface {
	SaveCachedMedia(ctx context.Context, item *media.CachedMediaItem) error

	GetCachedMedia(ctx context.Context, sessionID, messageID string) (*media.CachedMediaItem, error)

	UpdateCachedMedia(ctx context.Context, item *media.CachedMediaItem) error

	DeleteCachedMedia(ctx context.Context, sessionID, messageID string) error

	ListCachedMedia(ctx context.Context, req *media.ListCachedMediaRequest) (*media.ListCachedMediaResponse, error)

	ClearExpiredCache(ctx context.Context, sessionID string, olderThan time.Duration) (*media.ClearCacheResponse, error)

	GetMediaStats(ctx context.Context, sessionID string) (*media.MediaStats, error)

	CleanupOrphanedFiles(ctx context.Context, sessionID string) error

	GetCacheSize(ctx context.Context, sessionID string) (int64, error)

	GetCacheCount(ctx context.Context, sessionID string) (int, error)

	GetExpiredCacheItems(ctx context.Context, sessionID string) ([]*media.CachedMediaItem, error)

	BulkDeleteCachedMedia(ctx context.Context, sessionID string, messageIDs []string) error

	GetCachedMediaByType(ctx context.Context, sessionID string, mediaType string) ([]*media.CachedMediaItem, error)

	UpdateLastAccess(ctx context.Context, sessionID, messageID string) error

	GetLeastRecentlyUsed(ctx context.Context, sessionID string, limit int) ([]*media.CachedMediaItem, error)

	GetCacheSizeByType(ctx context.Context, sessionID string) (map[string]int64, error)

	GetCacheCountByType(ctx context.Context, sessionID string) (map[string]int, error)
}
