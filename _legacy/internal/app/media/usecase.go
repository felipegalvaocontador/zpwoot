package media

import (
	"context"
	"fmt"
	"time"

	"zpwoot/internal/domain/media"
	"zpwoot/internal/ports"
	"zpwoot/platform/logger"
)

const (
	defaultCacheDuration = 24 * time.Hour
)

type UseCase interface {
	DownloadMedia(ctx context.Context, req *DownloadMediaRequest) (*DownloadMediaResponse, error)
	GetMediaInfo(ctx context.Context, req *GetMediaInfoRequest) (*MediaInfoResponse, error)
	ListCachedMedia(ctx context.Context, req *ListCachedMediaRequest) (*ListCachedMediaResponse, error)
	ClearCache(ctx context.Context, req *ClearCacheRequest) (*ClearCacheResponse, error)
	GetMediaStats(ctx context.Context, req *GetMediaStatsRequest) (*GetMediaStatsResponse, error)
}

type useCaseImpl struct {
	mediaService media.Service
	mediaRepo    ports.MediaRepository
	logger       *logger.Logger
}

func NewUseCase(mediaService media.Service, mediaRepo ports.MediaRepository, appLogger *logger.Logger) UseCase {
	return &useCaseImpl{
		mediaService: mediaService,
		mediaRepo:    mediaRepo,
		logger:       appLogger,
	}
}

func (uc *useCaseImpl) DownloadMedia(ctx context.Context, req *DownloadMediaRequest) (*DownloadMediaResponse, error) {
	uc.logDownloadRequest(req)

	if response, served := uc.tryServeFromCache(ctx, req); served {
		return response, nil
	}

	return uc.downloadAndCacheMedia(ctx, req)
}

func (uc *useCaseImpl) logDownloadRequest(req *DownloadMediaRequest) {
	uc.logger.InfoWithFields("Downloading media", map[string]interface{}{
		"session_id": req.SessionID,
		"message_id": req.MessageID,
		"media_type": req.MediaType,
	})
}

func (uc *useCaseImpl) tryServeFromCache(ctx context.Context, req *DownloadMediaRequest) (*DownloadMediaResponse, bool) {
	cached, err := uc.mediaRepo.GetCachedMedia(ctx, req.SessionID, req.MessageID)
	if err != nil || cached == nil {
		return nil, false
	}

	if !time.Now().Before(cached.ExpiresAt) {
		return nil, false
	}

	uc.logger.InfoWithFields("Serving media from cache", map[string]interface{}{
		"session_id": req.SessionID,
		"message_id": req.MessageID,
		"file_path":  cached.FilePath,
	})

	uc.updateCacheAccessTime(ctx, cached, req)

	data, err := uc.mediaService.ReadCachedFile(ctx, cached.FilePath)
	if err != nil {
		uc.logger.WarnWithFields("Failed to read cached file, downloading fresh", map[string]interface{}{
			"session_id": req.SessionID,
			"message_id": req.MessageID,
			"error":      err.Error(),
		})
		return nil, false
	}

	return &DownloadMediaResponse{
		Data:     data,
		MimeType: cached.MimeType,
		FileSize: cached.FileSize,
		Filename: cached.Filename,
	}, true
}

func (uc *useCaseImpl) updateCacheAccessTime(ctx context.Context, cached *media.CachedMediaItem, req *DownloadMediaRequest) {
	cached.LastAccess = time.Now()
	if err := uc.mediaRepo.UpdateCachedMedia(ctx, cached); err != nil {
		uc.logger.WarnWithFields("Failed to update cached media access time", map[string]interface{}{
			"session_id": req.SessionID,
			"message_id": req.MessageID,
			"error":      err.Error(),
		})
	}
}

func (uc *useCaseImpl) downloadAndCacheMedia(ctx context.Context, req *DownloadMediaRequest) (*DownloadMediaResponse, error) {
	domainReq := &media.DownloadMediaRequest{
		SessionID: req.SessionID,
		MessageID: req.MessageID,
		MediaType: req.MediaType,
	}

	result, err := uc.mediaService.DownloadMedia(ctx, domainReq)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to download media", map[string]interface{}{
			"session_id": req.SessionID,
			"message_id": req.MessageID,
			"error":      err.Error(),
		})
		return nil, err
	}

	uc.cacheDownloadedMedia(ctx, req, result)

	return &DownloadMediaResponse{
		Data:     result.Data,
		MimeType: result.MimeType,
		FileSize: result.FileSize,
		Filename: result.Filename,
	}, nil
}

func (uc *useCaseImpl) cacheDownloadedMedia(ctx context.Context, req *DownloadMediaRequest, result *media.DownloadMediaResponse) {
	cacheItem := &media.CachedMediaItem{
		SessionID:  req.SessionID,
		MessageID:  req.MessageID,
		MediaType:  result.MediaType,
		MimeType:   result.MimeType,
		FileSize:   result.FileSize,
		Filename:   result.Filename,
		FilePath:   result.FilePath,
		CachedAt:   time.Now(),
		LastAccess: time.Now(),
		ExpiresAt:  time.Now().Add(defaultCacheDuration),
	}

	if err := uc.mediaRepo.SaveCachedMedia(ctx, cacheItem); err != nil {
		uc.logger.WarnWithFields("Failed to cache media", map[string]interface{}{
			"session_id": req.SessionID,
			"message_id": req.MessageID,
			"error":      err.Error(),
		})
	}
}

func (uc *useCaseImpl) GetMediaInfo(ctx context.Context, req *GetMediaInfoRequest) (*MediaInfoResponse, error) {
	uc.logger.InfoWithFields("Getting media info", map[string]interface{}{
		"session_id": req.SessionID,
		"message_id": req.MessageID,
	})

	domainReq := &media.GetMediaInfoRequest{
		SessionID: req.SessionID,
		MessageID: req.MessageID,
	}

	result, err := uc.mediaService.GetMediaInfo(ctx, domainReq)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to get media info", map[string]interface{}{
			"session_id": req.SessionID,
			"message_id": req.MessageID,
			"error":      err.Error(),
		})
		return nil, err
	}

	cached, err := uc.mediaRepo.GetCachedMedia(ctx, req.SessionID, req.MessageID)
	if err != nil {
		uc.logger.WarnWithFields("Failed to check cached media", map[string]interface{}{
			"session_id": req.SessionID,
			"message_id": req.MessageID,
			"error":      err.Error(),
		})
		cached = nil
	}
	isDownloaded := cached != nil && time.Now().Before(cached.ExpiresAt)
	var cacheExpiry time.Time
	if cached != nil {
		cacheExpiry = cached.ExpiresAt
	}

	return &MediaInfoResponse{
		MessageID:    result.MessageID,
		MediaType:    result.MediaType,
		MimeType:     result.MimeType,
		FileSize:     result.FileSize,
		Filename:     result.Filename,
		Caption:      result.Caption,
		Timestamp:    result.Timestamp,
		FromJID:      result.FromJID,
		IsDownloaded: isDownloaded,
		CacheExpiry:  cacheExpiry,
	}, nil
}

func (uc *useCaseImpl) ListCachedMedia(ctx context.Context, req *ListCachedMediaRequest) (*ListCachedMediaResponse, error) {
	uc.logger.InfoWithFields("Listing cached media", map[string]interface{}{
		"session_id": req.SessionID,
		"limit":      req.Limit,
		"offset":     req.Offset,
		"media_type": req.MediaType,
	})

	domainReq := &media.ListCachedMediaRequest{
		SessionID: req.SessionID,
		Limit:     req.Limit,
		Offset:    req.Offset,
		MediaType: req.MediaType,
	}

	result, err := uc.mediaService.ListCachedMedia(ctx, domainReq)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to list cached media", map[string]interface{}{
			"session_id": req.SessionID,
			"error":      err.Error(),
		})
		return nil, err
	}

	items := make([]CachedMediaItem, len(result.Items))
	for i, item := range result.Items {
		items[i] = CachedMediaItem{
			MessageID:  item.MessageID,
			MediaType:  item.MediaType,
			MimeType:   item.MimeType,
			FileSize:   item.FileSize,
			Filename:   item.Filename,
			CachedAt:   item.CachedAt,
			LastAccess: item.LastAccess,
			ExpiresAt:  item.ExpiresAt,
			FilePath:   item.FilePath,
		}
	}

	return &ListCachedMediaResponse{
		Items:     items,
		Total:     result.Total,
		Limit:     result.Limit,
		Offset:    result.Offset,
		HasMore:   result.HasMore,
		TotalSize: result.TotalSize,
	}, nil
}

func (uc *useCaseImpl) ClearCache(ctx context.Context, req *ClearCacheRequest) (*ClearCacheResponse, error) {
	uc.logger.InfoWithFields("Clearing media cache", map[string]interface{}{
		"session_id": req.SessionID,
		"older_than": req.OlderThan,
		"media_type": req.MediaType,
	})

	domainReq := &media.ClearCacheRequest{
		SessionID: req.SessionID,
		OlderThan: req.OlderThan,
		MediaType: req.MediaType,
	}

	result, err := uc.mediaService.ClearCache(ctx, domainReq)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to clear media cache", map[string]interface{}{
			"session_id": req.SessionID,
			"error":      err.Error(),
		})
		return nil, err
	}

	message := fmt.Sprintf("Successfully cleared %d files, freed %d bytes", result.FilesDeleted, result.SpaceFreed)

	return &ClearCacheResponse{
		FilesDeleted: result.FilesDeleted,
		SpaceFreed:   result.SpaceFreed,
		Message:      message,
	}, nil
}

func (uc *useCaseImpl) GetMediaStats(ctx context.Context, req *GetMediaStatsRequest) (*GetMediaStatsResponse, error) {
	uc.logger.InfoWithFields("Getting media stats", map[string]interface{}{
		"session_id": req.SessionID,
	})

	domainReq := &media.GetMediaStatsRequest{
		SessionID: req.SessionID,
	}

	result, err := uc.mediaService.GetMediaStats(ctx, domainReq)
	if err != nil {
		uc.logger.ErrorWithFields("Failed to get media stats", map[string]interface{}{
			"session_id": req.SessionID,
			"error":      err.Error(),
		})
		return nil, err
	}

	stats := MediaStats{
		TotalFiles:    result.Stats.TotalFiles,
		TotalSize:     result.Stats.TotalSize,
		ImageFiles:    result.Stats.ImageFiles,
		VideoFiles:    result.Stats.VideoFiles,
		AudioFiles:    result.Stats.AudioFiles,
		DocumentFiles: result.Stats.DocumentFiles,
		CacheHitRate:  result.Stats.CacheHitRate,
		AvgFileSize:   result.Stats.AvgFileSize,
	}

	return &GetMediaStatsResponse{
		SessionID: req.SessionID,
		Stats:     stats,
		UpdatedAt: result.UpdatedAt,
	}, nil
}
