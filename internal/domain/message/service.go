package message

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"zpwoot/platform/logger"
)

type MediaProcessor struct {
	logger  *logger.Logger
	tempDir string
	maxSize int64
	timeout time.Duration
}

func NewMediaProcessor(logger *logger.Logger) *MediaProcessor {
	return &MediaProcessor{
		logger:  logger,
		tempDir: os.TempDir(),
		maxSize: 100 * 1024 * 1024,
		timeout: 60 * time.Second,
	}
}

func (mp *MediaProcessor) ProcessMediaForType(ctx context.Context, file string, messageType MessageType) (*ProcessedMedia, error) {
	media, err := mp.ProcessMedia(ctx, file)
	if err != nil {
		return nil, err
	}

	if err := mp.validateMediaForType(media, messageType); err != nil {
		if media.Cleanup != nil {
			if cleanupErr := media.Cleanup(); cleanupErr != nil {
				mp.logger.WarnWithFields("Failed to cleanup media after validation error", map[string]interface{}{
					"cleanup_error":    cleanupErr.Error(),
					"validation_error": err.Error(),
				})
			}
		}
		return nil, err
	}

	return media, nil
}

func (mp *MediaProcessor) validateMediaForType(media *ProcessedMedia, messageType MessageType) error {
	switch messageType {
	case MessageTypeSticker:
		if !strings.Contains(media.MimeType, "webp") {
			return fmt.Errorf("stickers must be WebP format, got: %s", media.MimeType)
		}
		if media.FileSize > 100*1024 {
			return fmt.Errorf("sticker size exceeds 100KB limit: %d bytes", media.FileSize)
		}
		mp.logger.InfoWithFields("Sticker validation passed", map[string]interface{}{
			"mime_type": media.MimeType,
			"file_size": media.FileSize,
		})
	case MessageTypeImage:
		if media.FileSize > 10*1024*1024 {
			mp.logger.WarnWithFields("Large image file", map[string]interface{}{
				"file_size": media.FileSize,
			})
		}
	case MessageTypeVideo:
		if media.FileSize > 50*1024*1024 {
			mp.logger.WarnWithFields("Large video file", map[string]interface{}{
				"file_size": media.FileSize,
			})
		}
	}
	return nil
}

type ProcessedMedia struct {
	Cleanup  func() error
	FilePath string
	MimeType string
	FileSize int64
}

func (mp *MediaProcessor) ProcessMedia(ctx context.Context, file string) (*ProcessedMedia, error) {
	if file == "" {
		return nil, fmt.Errorf("file content is empty")
	}

	if strings.HasPrefix(file, "data:") {
		return mp.processBase64(file)
	}

	if strings.HasPrefix(file, "http://") || strings.HasPrefix(file, "https://") {
		return mp.processURL(ctx, file)
	}

	return nil, fmt.Errorf("unsupported file format: must be URL or base64")
}

func (mp *MediaProcessor) processBase64(data string) (*ProcessedMedia, error) {
	mp.logger.Debug("Processing base64 media")

	parts := strings.SplitN(data, ",", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid base64 data format")
	}

	mimeType := "application/octet-stream"
	if strings.Contains(parts[0], ":") && strings.Contains(parts[0], ";") {
		mimePart := strings.Split(parts[0], ":")[1]
		mimeType = strings.Split(mimePart, ";")[0]
	}

	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	if int64(len(decoded)) > mp.maxSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %d bytes", mp.maxSize)
	}

	tempFile, err := os.CreateTemp(mp.tempDir, "whatsmeow-media-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	if _, err := tempFile.Write(decoded); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
		return nil, fmt.Errorf("failed to write data to temporary file: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempFile.Name())
		return nil, fmt.Errorf("failed to close temporary file: %w", err)
	}

	mp.logger.InfoWithFields("Base64 media processed", map[string]interface{}{
		"file_path": tempFile.Name(),
		"mime_type": mimeType,
		"file_size": len(decoded),
	})

	return &ProcessedMedia{
		FilePath: tempFile.Name(),
		MimeType: mimeType,
		FileSize: int64(len(decoded)),
		Cleanup: func() error {
			return os.Remove(tempFile.Name())
		},
	}, nil
}

func (mp *MediaProcessor) processURL(ctx context.Context, url string) (*ProcessedMedia, error) {
	mp.logURLProcessing(url)

	resp, err := mp.downloadFromURL(ctx, url)
	if err != nil {
		return nil, err
	}
	defer mp.closeResponse(resp)

	mimeType, err := mp.validateResponse(resp)
	if err != nil {
		return nil, err
	}

	return mp.saveToTempFile(resp, url, mimeType)
}

func (mp *MediaProcessor) logURLProcessing(url string) {
	mp.logger.InfoWithFields("Processing URL media", map[string]interface{}{
		"url": url,
	})
}

func (mp *MediaProcessor) downloadFromURL(ctx context.Context, url string) (*http.Response, error) {
	client := &http.Client{Timeout: mp.timeout}

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("User-Agent", "zpwoot/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file from URL: %w", err)
	}

	return resp, nil
}

func (mp *MediaProcessor) closeResponse(resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		mp.logger.WarnWithFields("Failed to close response body", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

func (mp *MediaProcessor) validateResponse(resp *http.Response) (string, error) {
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download file: HTTP %d", resp.StatusCode)
	}

	if resp.ContentLength > mp.maxSize {
		return "", fmt.Errorf("file size exceeds maximum allowed size of %d bytes", mp.maxSize)
	}

	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return mimeType, nil
}

func (mp *MediaProcessor) saveToTempFile(resp *http.Response, url, mimeType string) (*ProcessedMedia, error) {
	tempFile, err := os.CreateTemp(mp.tempDir, "whatsmeow-media-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	written, err := io.CopyN(tempFile, resp.Body, mp.maxSize+1)
	if err != nil && !errors.Is(err, io.EOF) {
		mp.cleanupTempFile(tempFile)
		return nil, fmt.Errorf("failed to copy data to temporary file: %w", err)
	}

	if written > mp.maxSize {
		mp.cleanupTempFile(tempFile)
		return nil, fmt.Errorf("file size exceeds maximum allowed size of %d bytes", mp.maxSize)
	}

	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempFile.Name())
		return nil, fmt.Errorf("failed to close temporary file: %w", err)
	}

	mp.logURLProcessingSuccess(url, tempFile.Name(), mimeType, written)

	return &ProcessedMedia{
		FilePath: tempFile.Name(),
		MimeType: mimeType,
		FileSize: written,
		Cleanup: func() error {
			return os.Remove(tempFile.Name())
		},
	}, nil
}

func (mp *MediaProcessor) cleanupTempFile(tempFile *os.File) {
	_ = tempFile.Close()
	_ = os.Remove(tempFile.Name())
}

func (mp *MediaProcessor) logURLProcessingSuccess(url, filePath, mimeType string, fileSize int64) {
	mp.logger.InfoWithFields("URL media processed", map[string]interface{}{
		"url":       url,
		"file_path": filePath,
		"mime_type": mimeType,
		"file_size": fileSize,
	})
}

func DetectMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".mp4":  "video/mp4",
		".avi":  "video/avi",
		".mov":  "video/quicktime",
		".webm": "video/webm",
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".ogg":  "audio/ogg",
		".m4a":  "audio/mp4",
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".txt":  "text/plain",
	}

	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}

	return "application/octet-stream"
}

func ValidateMessageRequest(req *SendMessageRequest) error {
	if req.To == "" {
		return fmt.Errorf("recipient (to) is required")
	}

	if req.Type == "" {
		return fmt.Errorf("message type is required")
	}

	switch req.Type {
	case MessageTypeText:
		if req.Body == "" {
			return fmt.Errorf("body is required for text messages")
		}
	case MessageTypeImage, MessageTypeAudio, MessageTypeVideo, MessageTypeDocument, MessageTypeSticker:
		if req.File == "" {
			return fmt.Errorf("file is required for %s messages", req.Type)
		}
	case MessageTypeLocation:
		if req.Latitude == 0 || req.Longitude == 0 {
			return fmt.Errorf("latitude and longitude are required for location messages")
		}
	case MessageTypeContact:
		if req.ContactName == "" || req.ContactPhone == "" {
			return fmt.Errorf("contact name and phone are required for contact messages")
		}
	default:
		return fmt.Errorf("unsupported message type: %s", req.Type)
	}

	return nil
}
