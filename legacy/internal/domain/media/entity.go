package media

import (
	"time"
)

type MediaType string

const (
	MediaTypeImage    MediaType = "image"
	MediaTypeVideo    MediaType = "video"
	MediaTypeAudio    MediaType = "audio"
	MediaTypeDocument MediaType = "document"
	MediaTypeSticker  MediaType = "sticker"
)

type DownloadMediaRequest struct {
	SessionID string
	MessageID string
	MediaType string
}

type DownloadMediaResponse struct {
	MimeType  string
	Filename  string
	MediaType string
	FilePath  string
	Data      []byte
	FileSize  int64
}

type GetMediaInfoRequest struct {
	SessionID string
	MessageID string
}

type MediaInfo struct {
	Timestamp time.Time
	MessageID string
	MediaType string
	MimeType  string
	Filename  string
	Caption   string
	FromJID   string
	FileSize  int64
}

type CachedMediaItem struct {
	CachedAt   time.Time
	LastAccess time.Time
	ExpiresAt  time.Time
	SessionID  string
	MessageID  string
	MediaType  string
	MimeType   string
	Filename   string
	FilePath   string
	FileSize   int64
}

type ListCachedMediaRequest struct {
	SessionID string
	MediaType string
	Limit     int
	Offset    int
}

type ListCachedMediaResponse struct {
	Items     []CachedMediaItem
	Total     int
	Limit     int
	Offset    int
	HasMore   bool
	TotalSize int64
}

type ClearCacheRequest struct {
	SessionID string
	MediaType string
	OlderThan int
}

type ClearCacheResponse struct {
	FilesDeleted int
	SpaceFreed   int64
}

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

type GetMediaStatsRequest struct {
	SessionID string
}

type GetMediaStatsResponse struct {
	UpdatedAt time.Time
	SessionID string
	Stats     MediaStats
}

func IsValidMediaType(mediaType string) bool {
	switch MediaType(mediaType) {
	case MediaTypeImage, MediaTypeVideo, MediaTypeAudio, MediaTypeDocument, MediaTypeSticker:
		return true
	default:
		return false
	}
}

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

func GetMediaTypeFromMimeType(mimeType string) MediaType {
	switch {
	case mimeType == "image/webp":
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

func ValidateMediaInfoRequest(req *GetMediaInfoRequest) error {
	if req.SessionID == "" {
		return ErrInvalidSessionID
	}
	if req.MessageID == "" {
		return ErrInvalidMessageID
	}
	return nil
}

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
