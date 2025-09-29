package media

import "time"

// DownloadMediaRequest represents a request to download media from a message
type DownloadMediaRequest struct {
	SessionID string `json:"sessionId" validate:"required" example:"session-123"`
	MessageID string `json:"messageId" validate:"required" example:"3EB0C431C26A1916E07E"`
	MediaType string `json:"mediaType,omitempty" example:"image"` // Optional filter by media type
}

// DownloadMediaResponse represents the response containing downloaded media
type DownloadMediaResponse struct {
	Data     []byte `json:"-"`                                      // Binary data (not serialized in JSON)
	MimeType string `json:"mimeType" example:"image/jpeg"`          // MIME type of the media
	FileSize int64  `json:"fileSize" example:"1024000"`             // File size in bytes
	Filename string `json:"filename,omitempty" example:"image.jpg"` // Original filename if available
}

// GetMediaInfoRequest represents a request to get media information
type GetMediaInfoRequest struct {
	SessionID string `json:"sessionId" validate:"required" example:"session-123"`
	MessageID string `json:"messageId" validate:"required" example:"3EB0C431C26A1916E07E"`
}

// MediaInfoResponse represents media information without the actual data
type MediaInfoResponse struct {
	MessageID    string    `json:"messageId" example:"3EB0C431C26A1916E07E"`
	MediaType    string    `json:"mediaType" example:"image"`
	MimeType     string    `json:"mimeType" example:"image/jpeg"`
	FileSize     int64     `json:"fileSize" example:"1024000"`
	Filename     string    `json:"filename,omitempty" example:"image.jpg"`
	Caption      string    `json:"caption,omitempty" example:"Beautiful sunset"`
	Timestamp    time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
	FromJID      string    `json:"fromJid" example:"5511999999999@s.whatsapp.net"`
	IsDownloaded bool      `json:"isDownloaded" example:"true"`
	CacheExpiry  time.Time `json:"cacheExpiry,omitempty" example:"2024-01-02T12:00:00Z"`
}

// ListCachedMediaRequest represents a request to list cached media
type ListCachedMediaRequest struct {
	SessionID string `json:"sessionId" validate:"required" example:"session-123"`
	Limit     int    `json:"limit" validate:"min=1,max=100" example:"50"`
	Offset    int    `json:"offset" validate:"min=0" example:"0"`
	MediaType string `json:"media_type,omitempty" example:"image"` // Optional filter by media type
}

// CachedMediaItem represents a single cached media item
type CachedMediaItem struct {
	MessageID  string    `json:"messageId" example:"3EB0C431C26A1916E07E"`
	MediaType  string    `json:"mediaType" example:"image"`
	MimeType   string    `json:"mimeType" example:"image/jpeg"`
	FileSize   int64     `json:"fileSize" example:"1024000"`
	Filename   string    `json:"filename,omitempty" example:"image.jpg"`
	CachedAt   time.Time `json:"cachedAt" example:"2024-01-01T12:00:00Z"`
	LastAccess time.Time `json:"lastAccess" example:"2024-01-01T12:30:00Z"`
	ExpiresAt  time.Time `json:"expiresAt" example:"2024-01-02T12:00:00Z"`
	FilePath   string    `json:"filePath,omitempty" example:"/tmp/media/abc123.jpg"`
}

// ListCachedMediaResponse represents the response for listing cached media
type ListCachedMediaResponse struct {
	Items     []CachedMediaItem `json:"items"`
	Total     int               `json:"total" example:"150"`
	Limit     int               `json:"limit" example:"50"`
	Offset    int               `json:"offset" example:"0"`
	HasMore   bool              `json:"hasMore" example:"true"`
	TotalSize int64             `json:"totalSize" example:"52428800"` // Total size in bytes
}

// ClearCacheRequest represents a request to clear media cache
type ClearCacheRequest struct {
	SessionID string `json:"sessionId" validate:"required" example:"session-123"`
	OlderThan int    `json:"olderThan" validate:"min=0" example:"24"` // Hours
	MediaType string `json:"mediaType,omitempty" example:"image"`     // Optional filter by media type
}

// ClearCacheResponse represents the response for clearing cache
type ClearCacheResponse struct {
	FilesDeleted int    `json:"filesDeleted" example:"25"`
	SpaceFreed   int64  `json:"spaceFreed" example:"10485760"` // Bytes freed
	Message      string `json:"message" example:"Cache cleared successfully"`
}

// MediaStats represents statistics about media usage
type MediaStats struct {
	TotalFiles    int     `json:"totalFiles" example:"100"`
	TotalSize     int64   `json:"totalSize" example:"52428800"`
	ImageFiles    int     `json:"imageFiles" example:"60"`
	VideoFiles    int     `json:"videoFiles" example:"20"`
	AudioFiles    int     `json:"audioFiles" example:"15"`
	DocumentFiles int     `json:"documentFiles" example:"5"`
	CacheHitRate  float64 `json:"cacheHitRate" example:"0.85"`
	AvgFileSize   int64   `json:"avgFileSize" example:"524288"`
}

// GetMediaStatsRequest represents a request to get media statistics
type GetMediaStatsRequest struct {
	SessionID string `json:"sessionId" validate:"required" example:"session-123"`
}

// GetMediaStatsResponse represents the response for media statistics
type GetMediaStatsResponse struct {
	SessionID string     `json:"sessionId" example:"session-123"`
	Stats     MediaStats `json:"stats"`
	UpdatedAt time.Time  `json:"updatedAt" example:"2024-01-01T12:00:00Z"`
}
