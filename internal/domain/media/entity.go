package media

import (
	"time"
)

// MediaType represents different types of media
type MediaType string

const (
	MediaTypeImage    MediaType = "image"
	MediaTypeVideo    MediaType = "video"
	MediaTypeAudio    MediaType = "audio"
	MediaTypeDocument MediaType = "document"
	MediaTypeSticker  MediaType = "sticker"
)

// DownloadMediaRequest represents a request to download media
type DownloadMediaRequest struct {
	SessionID string
	MessageID string
	MediaType string // Optional filter
}

// DownloadMediaResponse represents the result of downloading media
type DownloadMediaResponse struct {
	Data      []byte
	MimeType  string
	FileSize  int64
	Filename  string
	MediaType string
	FilePath  string // Path where file is cached
}

// GetMediaInfoRequest represents a request to get media information
type GetMediaInfoRequest struct {
	SessionID string
	MessageID string
}

// MediaInfo represents information about media in a message
type MediaInfo struct {
	MessageID string
	MediaType string
	MimeType  string
	FileSize  int64
	Filename  string
	Caption   string
	Timestamp time.Time
	FromJID   string
}

// CachedMediaItem represents a cached media file
type CachedMediaItem struct {
	SessionID  string
	MessageID  string
	MediaType  string
	MimeType   string
	FileSize   int64
	Filename   string
	FilePath   string
	CachedAt   time.Time
	LastAccess time.Time
	ExpiresAt  time.Time
}

// ListCachedMediaRequest represents a request to list cached media
type ListCachedMediaRequest struct {
	SessionID string
	Limit     int
	Offset    int
	MediaType string // Optional filter
}

// ListCachedMediaResponse represents the result of listing cached media
type ListCachedMediaResponse struct {
	Items     []CachedMediaItem
	Total     int
	Limit     int
	Offset    int
	HasMore   bool
	TotalSize int64
}

// ClearCacheRequest represents a request to clear media cache
type ClearCacheRequest struct {
	SessionID string
	OlderThan int    // Hours
	MediaType string // Optional filter
}

// ClearCacheResponse represents the result of clearing cache
type ClearCacheResponse struct {
	FilesDeleted int
	SpaceFreed   int64
}

// MediaStats represents statistics about media usage
type MediaStats struct {
	TotalFiles    int
	TotalSize     int64
	ImageFiles    int
	VideoFiles    int
	AudioFiles    int
	DocumentFiles int
	CacheHitRate  float64
	AvgFileSize   int64
}

// GetMediaStatsRequest represents a request to get media statistics
type GetMediaStatsRequest struct {
	SessionID string
}

// GetMediaStatsResponse represents media statistics
type GetMediaStatsResponse struct {
	SessionID string
	Stats     MediaStats
	UpdatedAt time.Time
}

// IsValidMediaType checks if a media type is valid
func IsValidMediaType(mediaType string) bool {
	switch MediaType(mediaType) {
	case MediaTypeImage, MediaTypeVideo, MediaTypeAudio, MediaTypeDocument, MediaTypeSticker:
		return true
	default:
		return false
	}
}

// GetMimeTypeForMediaType returns common MIME types for a media type
func GetMimeTypeForMediaType(mediaType MediaType) []string {
	switch mediaType {
	case MediaTypeImage:
		return []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
	case MediaTypeVideo:
		return []string{"video/mp4", "video/avi", "video/mov", "video/webm"}
	case MediaTypeAudio:
		return []string{"audio/ogg", "audio/mp3", "audio/wav", "audio/aac"}
	case MediaTypeDocument:
		return []string{"application/pdf", "application/msword", "text/plain", "application/octet-stream"}
	case MediaTypeSticker:
		return []string{"image/webp"}
	default:
		return []string{"application/octet-stream"}
	}
}

// GetMediaTypeFromMimeType determines media type from MIME type
func GetMediaTypeFromMimeType(mimeType string) MediaType {
	switch {
	case mimeType == "image/webp":
		// WebP could be sticker or image, default to image
		return MediaTypeImage
	case len(mimeType) >= 5 && mimeType[:5] == "image":
		return MediaTypeImage
	case len(mimeType) >= 5 && mimeType[:5] == "video":
		return MediaTypeVideo
	case len(mimeType) >= 5 && mimeType[:5] == "audio":
		return MediaTypeAudio
	default:
		return MediaTypeDocument
	}
}

// ValidateDownloadRequest validates a download media request
func ValidateDownloadRequest(req *DownloadMediaRequest) error {
	if req.SessionID == "" {
		return ErrInvalidSessionID
	}
	if req.MessageID == "" {
		return ErrInvalidMessageID
	}
	if req.MediaType != "" && !IsValidMediaType(req.MediaType) {
		return ErrInvalidMediaType
	}
	return nil
}

// ValidateMediaInfoRequest validates a get media info request
func ValidateMediaInfoRequest(req *GetMediaInfoRequest) error {
	if req.SessionID == "" {
		return ErrInvalidSessionID
	}
	if req.MessageID == "" {
		return ErrInvalidMessageID
	}
	return nil
}

// ValidateListCachedMediaRequest validates a list cached media request
func ValidateListCachedMediaRequest(req *ListCachedMediaRequest) error {
	if req.SessionID == "" {
		return ErrInvalidSessionID
	}
	if req.Limit <= 0 || req.Limit > 100 {
		return ErrInvalidLimit
	}
	if req.Offset < 0 {
		return ErrInvalidOffset
	}
	if req.MediaType != "" && !IsValidMediaType(req.MediaType) {
		return ErrInvalidMediaType
	}
	return nil
}

// ValidateClearCacheRequest validates a clear cache request
func ValidateClearCacheRequest(req *ClearCacheRequest) error {
	if req.SessionID == "" {
		return ErrInvalidSessionID
	}
	if req.OlderThan < 0 {
		return ErrInvalidOlderThan
	}
	if req.MediaType != "" && !IsValidMediaType(req.MediaType) {
		return ErrInvalidMediaType
	}
	return nil
}
