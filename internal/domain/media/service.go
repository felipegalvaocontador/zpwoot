package media

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"time"

	"zpwoot/platform/logger"
)

// Service defines the interface for media domain service
type Service interface {
	DownloadMedia(ctx context.Context, req *DownloadMediaRequest) (*DownloadMediaResponse, error)
	GetMediaInfo(ctx context.Context, req *GetMediaInfoRequest) (*MediaInfo, error)
	ListCachedMedia(ctx context.Context, req *ListCachedMediaRequest) (*ListCachedMediaResponse, error)
	ClearCache(ctx context.Context, req *ClearCacheRequest) (*ClearCacheResponse, error)
	GetMediaStats(ctx context.Context, req *GetMediaStatsRequest) (*GetMediaStatsResponse, error)
	ReadCachedFile(ctx context.Context, filePath string) ([]byte, error)
}

// WhatsAppClient defines the interface for WhatsApp operations
type WhatsAppClient interface {
	IsLoggedIn() bool
	DownloadMedia(ctx context.Context, messageID string) ([]byte, string, error) // returns data, mimeType, error
	GetMessageInfo(ctx context.Context, messageID string) (*MessageInfo, error)
}

// MessageInfo represents information about a WhatsApp message
type MessageInfo struct {
	Timestamp time.Time
	ID        string
	FromJID   string
	MediaType string
	MimeType  string
	Filename  string
	Caption   string
	FileSize  int64
	HasMedia  bool
}

// CacheManager defines the interface for cache operations
type CacheManager interface {
	SaveFile(ctx context.Context, data []byte, filename string) (string, error)
	ReadFile(ctx context.Context, filePath string) ([]byte, error)
	DeleteFile(ctx context.Context, filePath string) error
	ListFiles(ctx context.Context, pattern string) ([]string, error)
	GetFileInfo(ctx context.Context, filePath string) (os.FileInfo, error)
	CleanupOldFiles(ctx context.Context, olderThan time.Duration) (int, int64, error)
}

type serviceImpl struct {
	whatsappClient WhatsAppClient
	cacheManager   CacheManager
	logger         *logger.Logger
	cacheDir       string
	maxFileSize    int64
}

// NewService creates a new media domain service
func NewService(whatsappClient WhatsAppClient, cacheManager CacheManager, logger *logger.Logger, cacheDir string) Service {
	return &serviceImpl{
		whatsappClient: whatsappClient,
		cacheManager:   cacheManager,
		logger:         logger,
		cacheDir:       cacheDir,
		maxFileSize:    100 * 1024 * 1024, // 100MB default
	}
}

// DownloadMedia downloads media from a WhatsApp message
func (s *serviceImpl) DownloadMedia(ctx context.Context, req *DownloadMediaRequest) (*DownloadMediaResponse, error) {
	if err := ValidateDownloadRequest(req); err != nil {
		return nil, err
	}

	s.logDownloadRequest(req)

	// Validate client and message
	msgInfo, err := s.validateAndGetMessageInfo(ctx, req)
	if err != nil {
		return nil, err
	}

	// Download and process media
	return s.downloadAndProcessMedia(ctx, req, msgInfo)
}

// logDownloadRequest logs the download request
func (s *serviceImpl) logDownloadRequest(req *DownloadMediaRequest) {
	s.logger.InfoWithFields("Downloading media from WhatsApp", map[string]interface{}{
		"session_id": req.SessionID,
		"message_id": req.MessageID,
		"media_type": req.MediaType,
	})
}

// validateAndGetMessageInfo validates client and gets message info
func (s *serviceImpl) validateAndGetMessageInfo(ctx context.Context, req *DownloadMediaRequest) (*MessageInfo, error) {
	// Check if WhatsApp client is logged in
	if !s.whatsappClient.IsLoggedIn() {
		return nil, ErrClientNotLoggedIn
	}

	// Get message info first to validate media type if specified
	msgInfo, err := s.whatsappClient.GetMessageInfo(ctx, req.MessageID)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get message info", map[string]interface{}{
			"message_id": req.MessageID,
			"error":      err.Error(),
		})
		return nil, ErrMessageNotFound
	}

	if !msgInfo.HasMedia {
		return nil, ErrNoMediaInMessage
	}

	// Validate media type if specified
	if req.MediaType != "" && msgInfo.MediaType != req.MediaType {
		return nil, ErrMediaTypeMismatch
	}

	return msgInfo, nil
}

// downloadAndProcessMedia downloads media and processes it
func (s *serviceImpl) downloadAndProcessMedia(ctx context.Context, req *DownloadMediaRequest, msgInfo *MessageInfo) (*DownloadMediaResponse, error) {
	// Download media from WhatsApp
	data, mimeType, err := s.whatsappClient.DownloadMedia(ctx, req.MessageID)
	if err != nil {
		s.logger.ErrorWithFields("Failed to download media", map[string]interface{}{
			"message_id": req.MessageID,
			"error":      err.Error(),
		})
		return nil, ErrDownloadFailed
	}

	if int64(len(data)) > s.maxFileSize {
		return nil, ErrFileTooLarge
	}

	// Generate filename and cache
	filename := s.generateFilename(req.MessageID, mimeType, msgInfo.Filename)
	filePath := s.cacheMediaFile(ctx, data, filename, req.MessageID)

	s.logDownloadSuccess(req.MessageID, data, mimeType, filePath)

	return &DownloadMediaResponse{
		Data:      data,
		MimeType:  mimeType,
		FileSize:  int64(len(data)),
		Filename:  filename,
		MediaType: msgInfo.MediaType,
		FilePath:  filePath,
	}, nil
}

// cacheMediaFile caches the media file
func (s *serviceImpl) cacheMediaFile(ctx context.Context, data []byte, filename, messageID string) string {
	filePath, err := s.cacheManager.SaveFile(ctx, data, filename)
	if err != nil {
		s.logger.WarnWithFields("Failed to cache media file", map[string]interface{}{
			"message_id": messageID,
			"error":      err.Error(),
		})
		return ""
	}
	return filePath
}

// logDownloadSuccess logs successful download
func (s *serviceImpl) logDownloadSuccess(messageID string, data []byte, mimeType, filePath string) {
	s.logger.InfoWithFields("Media downloaded successfully", map[string]interface{}{
		"message_id": messageID,
		"file_size":  len(data),
		"mime_type":  mimeType,
		"cached":     filePath != "",
	})
}

// GetMediaInfo gets information about media in a message without downloading it
func (s *serviceImpl) GetMediaInfo(ctx context.Context, req *GetMediaInfoRequest) (*MediaInfo, error) {
	if err := ValidateMediaInfoRequest(req); err != nil {
		return nil, err
	}

	s.logger.InfoWithFields("Getting media info", map[string]interface{}{
		"session_id": req.SessionID,
		"message_id": req.MessageID,
	})

	// Check if WhatsApp client is logged in
	if !s.whatsappClient.IsLoggedIn() {
		return nil, ErrClientNotLoggedIn
	}

	// Get message info
	msgInfo, err := s.whatsappClient.GetMessageInfo(ctx, req.MessageID)
	if err != nil {
		s.logger.ErrorWithFields("Failed to get message info", map[string]interface{}{
			"message_id": req.MessageID,
			"error":      err.Error(),
		})
		return nil, ErrMessageNotFound
	}

	if !msgInfo.HasMedia {
		return nil, ErrNoMediaInMessage
	}

	return &MediaInfo{
		MessageID: msgInfo.ID,
		MediaType: msgInfo.MediaType,
		MimeType:  msgInfo.MimeType,
		FileSize:  msgInfo.FileSize,
		Filename:  msgInfo.Filename,
		Caption:   msgInfo.Caption,
		Timestamp: msgInfo.Timestamp,
		FromJID:   msgInfo.FromJID,
	}, nil
}

// ListCachedMedia lists cached media files
func (s *serviceImpl) ListCachedMedia(ctx context.Context, req *ListCachedMediaRequest) (*ListCachedMediaResponse, error) {
	if err := ValidateListCachedMediaRequest(req); err != nil {
		return nil, err
	}

	s.logListCachedMediaRequest(req)

	// Get cached files list
	files, err := s.getCachedFilesList(ctx, req.MediaType)
	if err != nil {
		return nil, err
	}

	// Apply pagination and build response
	return s.buildCachedMediaResponse(ctx, files, req)
}

// logListCachedMediaRequest logs the list cached media request
func (s *serviceImpl) logListCachedMediaRequest(req *ListCachedMediaRequest) {
	s.logger.InfoWithFields("Listing cached media", map[string]interface{}{
		"session_id": req.SessionID,
		"limit":      req.Limit,
		"offset":     req.Offset,
		"media_type": req.MediaType,
	})
}

// getCachedFilesList gets the list of cached files based on media type filter
func (s *serviceImpl) getCachedFilesList(ctx context.Context, mediaType string) ([]string, error) {
	// This is a simplified implementation
	// In a real implementation, you would query the cache database/storage
	pattern := "*"
	if mediaType != "" {
		pattern = fmt.Sprintf("*_%s_*", mediaType)
	}

	files, err := s.cacheManager.ListFiles(ctx, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to list cached files: %w", err)
	}

	return files, nil
}

// buildCachedMediaResponse builds the cached media response with pagination
func (s *serviceImpl) buildCachedMediaResponse(ctx context.Context, files []string, req *ListCachedMediaRequest) (*ListCachedMediaResponse, error) {
	total := len(files)
	start, end := s.calculatePaginationBounds(total, req.Offset, req.Limit)

	// Handle empty result case
	if start >= total {
		return s.buildEmptyResponse(req, total), nil
	}

	// Build items for the current page
	items, totalSize, err := s.buildCachedMediaItems(ctx, files[start:end])
	if err != nil {
		return nil, err
	}

	return &ListCachedMediaResponse{
		Items:     items,
		Total:     total,
		Limit:     req.Limit,
		Offset:    req.Offset,
		HasMore:   end < total,
		TotalSize: totalSize,
	}, nil
}

// calculatePaginationBounds calculates start and end indices for pagination
func (s *serviceImpl) calculatePaginationBounds(total, offset, limit int) (int, int) {
	start := offset
	end := start + limit

	if end > total {
		end = total
	}

	return start, end
}

// buildEmptyResponse builds an empty response for cases with no results
func (s *serviceImpl) buildEmptyResponse(req *ListCachedMediaRequest, total int) *ListCachedMediaResponse {
	return &ListCachedMediaResponse{
		Items:     []CachedMediaItem{},
		Total:     total,
		Limit:     req.Limit,
		Offset:    req.Offset,
		HasMore:   false,
		TotalSize: 0,
	}
}

// buildCachedMediaItems builds cached media items from file paths
func (s *serviceImpl) buildCachedMediaItems(ctx context.Context, filePaths []string) ([]CachedMediaItem, int64, error) {
	items := make([]CachedMediaItem, 0, len(filePaths))
	var totalSize int64

	for _, filePath := range filePaths {
		item, size, err := s.buildCachedMediaItem(ctx, filePath)
		if err != nil {
			// Skip files that can't be processed
			continue
		}

		items = append(items, item)
		totalSize += size
	}

	return items, totalSize, nil
}

// buildCachedMediaItem builds a single cached media item from file path
func (s *serviceImpl) buildCachedMediaItem(ctx context.Context, filePath string) (CachedMediaItem, int64, error) {
	info, err := s.cacheManager.GetFileInfo(ctx, filePath)
	if err != nil {
		return CachedMediaItem{}, 0, err
	}

	// Parse filename to extract metadata (simplified)
	filename := filepath.Base(filePath)

	item := CachedMediaItem{
		MessageID:  extractMessageIDFromFilename(filename),
		MediaType:  extractMediaTypeFromFilename(filename),
		MimeType:   extractMimeTypeFromFilename(filename),
		FileSize:   info.Size(),
		Filename:   filename,
		CachedAt:   info.ModTime(),
		LastAccess: info.ModTime(), // Simplified
		ExpiresAt:  info.ModTime().Add(24 * time.Hour),
		FilePath:   filePath,
	}

	return item, info.Size(), nil
}

// ClearCache clears cached media files
func (s *serviceImpl) ClearCache(ctx context.Context, req *ClearCacheRequest) (*ClearCacheResponse, error) {
	if err := ValidateClearCacheRequest(req); err != nil {
		return nil, err
	}

	s.logger.InfoWithFields("Clearing media cache", map[string]interface{}{
		"session_id": req.SessionID,
		"older_than": req.OlderThan,
		"media_type": req.MediaType,
	})

	olderThan := time.Duration(req.OlderThan) * time.Hour
	filesDeleted, spaceFreed, err := s.cacheManager.CleanupOldFiles(ctx, olderThan)
	if err != nil {
		return nil, fmt.Errorf("failed to cleanup cache: %w", err)
	}

	s.logger.InfoWithFields("Cache cleared successfully", map[string]interface{}{
		"files_deleted": filesDeleted,
		"space_freed":   spaceFreed,
	})

	return &ClearCacheResponse{
		FilesDeleted: filesDeleted,
		SpaceFreed:   spaceFreed,
	}, nil
}

// GetMediaStats gets statistics about media usage
func (s *serviceImpl) GetMediaStats(ctx context.Context, req *GetMediaStatsRequest) (*GetMediaStatsResponse, error) {
	s.logger.InfoWithFields("Getting media stats", map[string]interface{}{
		"session_id": req.SessionID,
	})

	// This is a simplified implementation
	// In a real implementation, you would query actual statistics
	stats := MediaStats{
		TotalFiles:    0,
		TotalSize:     0,
		ImageFiles:    0,
		VideoFiles:    0,
		AudioFiles:    0,
		DocumentFiles: 0,
		CacheHitRate:  0.85,
		AvgFileSize:   524288, // 512KB
	}

	return &GetMediaStatsResponse{
		SessionID: req.SessionID,
		Stats:     stats,
		UpdatedAt: time.Now(),
	}, nil
}

// ReadCachedFile reads a cached file
func (s *serviceImpl) ReadCachedFile(ctx context.Context, filePath string) ([]byte, error) {
	return s.cacheManager.ReadFile(ctx, filePath)
}

// Helper functions

func (s *serviceImpl) generateFilename(messageID, mimeType, originalFilename string) string {
	if originalFilename != "" {
		return originalFilename
	}

	// Generate filename based on message ID and MIME type
	ext := ""
	if mimeType != "" {
		exts, _ := mime.ExtensionsByType(mimeType)
		if len(exts) > 0 {
			ext = exts[0]
		}
	}

	return fmt.Sprintf("%s%s", messageID, ext)
}

// Simplified helper functions for filename parsing
func extractMessageIDFromFilename(filename string) string {
	// This is a simplified implementation
	// In a real implementation, you would parse the filename properly
	return filename
}

func extractMediaTypeFromFilename(filename string) string {
	// Extract media type from filename extension
	if len(filename) == 0 {
		return "unknown"
	}

	// Find the last dot in the filename
	lastDot := -1
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			lastDot = i
			break
		}
	}

	if lastDot == -1 || lastDot == len(filename)-1 {
		return "unknown"
	}

	ext := filename[lastDot+1:]

	// Map common extensions to media types
	switch ext {
	case "jpg", "jpeg", "png", "gif", "webp":
		return "image"
	case "mp4", "avi", "mov", "mkv", "webm":
		return "video"
	case "mp3", "wav", "ogg", "m4a", "aac":
		return "audio"
	case "pdf", "doc", "docx", "txt", "xls", "xlsx":
		return "document"
	default:
		return "unknown"
	}
}

func extractMimeTypeFromFilename(filename string) string {
	// This is a simplified implementation
	ext := filepath.Ext(filename)
	return mime.TypeByExtension(ext)
}
