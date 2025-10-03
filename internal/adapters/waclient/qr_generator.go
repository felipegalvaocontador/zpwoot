package waclient

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image/png"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mdp/qrterminal/v3"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"

	"zpwoot/internal/core/session"
	"zpwoot/platform/logger"
)


type QRGenerator struct {
	logger *logger.Logger


	mu            sync.RWMutex
	qrCode        string
	qrCodeExpires time.Time
	isActive      bool

	ctx    context.Context
	cancel context.CancelFunc
}

func NewQRGenerator(logger *logger.Logger) *QRGenerator {
	return &QRGenerator{
		logger: logger,
	}
}


func (g *QRGenerator) StartQRLoop(ctx context.Context, qrChan <-chan whatsmeow.QRChannelItem, sessionName string) {
	g.mu.Lock()
	defer g.mu.Unlock()


	g.stopInternal()


	g.ctx, g.cancel = context.WithCancel(ctx)
	g.isActive = true


	go g.runQRLoop(qrChan, sessionName)
}


func (g *QRGenerator) Stop() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.stopInternal()
}

func (g *QRGenerator) stopInternal() {
	if g.cancel != nil {
		g.cancel()
	}
	g.isActive = false
	g.qrCode = ""
	g.qrCodeExpires = time.Time{}
}

func (g *QRGenerator) runQRLoop(qrChan <-chan whatsmeow.QRChannelItem, sessionName string) {
	defer func() {
		if r := recover(); r != nil {
			g.logger.ErrorWithFields("QR loop panic", map[string]interface{}{
				"session_name": sessionName,
				"error":        r,
			})
		}
		g.mu.Lock()
		g.isActive = false
		g.mu.Unlock()
	}()

	g.logger.InfoWithFields("QR loop started", map[string]interface{}{
		"session_name": sessionName,
	})

	for {
		select {
		case <-g.ctx.Done():
			g.logger.InfoWithFields("QR loop cancelled", map[string]interface{}{
				"session_name": sessionName,
			})
			return

		case evt, ok := <-qrChan:
			if !ok {
				g.logger.InfoWithFields("QR channel closed", map[string]interface{}{
					"session_name": sessionName,
				})
				return
			}

			switch evt.Event {
			case "code":
				g.handleQRCode(evt.Code, sessionName)
			case "timeout":
				g.handleQRTimeout(sessionName)
				return
			case "success":
				g.handleQRSuccess(sessionName)
				return
			}
		}
	}
}

func (g *QRGenerator) handleQRCode(code string, sessionName string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.qrCode = code
	g.qrCodeExpires = time.Now().Add(30 * time.Second)

	g.logger.InfoWithFields("QR code generated", map[string]interface{}{
		"session_name": sessionName,
		"expires_at":   g.qrCodeExpires,
	})
}

func (g *QRGenerator) handleQRTimeout(sessionName string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.logger.WarnWithFields("QR code timeout", map[string]interface{}{
		"session_name": sessionName,
	})

	g.qrCode = ""
	g.qrCodeExpires = time.Time{}
}

func (g *QRGenerator) handleQRSuccess(sessionName string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.logger.InfoWithFields("QR code scan successful", map[string]interface{}{
		"session_name": sessionName,
	})

	g.qrCode = ""
	g.qrCodeExpires = time.Time{}
}

func (g *QRGenerator) Generate(ctx context.Context, sessionName string) (*session.QRCodeResponse, error) {

	return nil, fmt.Errorf("QR code generation is handled by WhatsApp events")
}


func (g *QRGenerator) GetQRCode() (string, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.qrCode == "" {
		return "", false
	}

	if time.Now().After(g.qrCodeExpires) {
		return "", false
	}

	return g.qrCode, true
}


func (g *QRGenerator) IsActive() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.isActive
}


func (g *QRGenerator) GetQRCodeExpiry() time.Time {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.qrCodeExpires
}

func (g *QRGenerator) GenerateQRCode(data string) (string, error) {

	return data, nil
}

func (g *QRGenerator) GenerateQRCodeImage(data string) (string, error) {
	g.logger.DebugWithFields("Generating QR code image", map[string]interface{}{
		"data_length": len(data),
	})

	qr, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		return "", fmt.Errorf("failed to create QR code: %w", err)
	}

	qr.DisableBorder = false

	img := qr.Image(256)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("failed to encode QR code image: %w", err)
	}

	base64Image := base64.StdEncoding.EncodeToString(buf.Bytes())
	dataURI := fmt.Sprintf("data:image/png;base64,%s", base64Image)

	g.logger.DebugWithFields("QR code image generated", map[string]interface{}{
		"image_size": len(base64Image),
	})

	return dataURI, nil
}

func (g *QRGenerator) GenerateQRCodePNG(data string, size int) ([]byte, error) {
	if size <= 0 {
		size = 256
	}

	g.logger.DebugWithFields("Generating QR code PNG", map[string]interface{}{
		"data_length": len(data),
		"size":        size,
	})

	qr, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("failed to create QR code: %w", err)
	}

	qr.DisableBorder = false

	pngBytes, err := qr.PNG(size)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code PNG: %w", err)
	}

	g.logger.DebugWithFields("QR code PNG generated", map[string]interface{}{
		"bytes_size": len(pngBytes),
	})

	return pngBytes, nil
}

func (g *QRGenerator) ValidateQRCode(data string) bool {

	if len(data) < 10 {
		return false
	}

	if data[0] < '0' || data[0] > '9' {
		return false
	}

	atIndex := -1
	for i, char := range data {
		if char == '@' {
			atIndex = i
			break
		}
	}

	if atIndex == -1 || atIndex == 0 {
		return false
	}

	if atIndex >= len(data)-1 {
		return false
	}

	return true
}

func (g *QRGenerator) GetQRCodeInfo(data string) map[string]interface{} {
	info := map[string]interface{}{
		"valid":  g.ValidateQRCode(data),
		"length": len(data),
	}

	if !g.ValidateQRCode(data) {
		return info
	}

	atIndex := -1
	for i, char := range data {
		if char == '@' {
			atIndex = i
			break
		}
	}

	if atIndex > 0 {
		info["version"] = data[:atIndex]
		info["payload"] = data[atIndex+1:]
		info["payload_length"] = len(data[atIndex+1:])
	}

	return info
}

func (g *QRGenerator) GenerateImage(ctx context.Context, qrCode string) ([]byte, error) {
	return g.GenerateQRCodePNG(qrCode, 256)
}

func (g *QRGenerator) IsExpired(expiresAt time.Time) bool {
	return time.Now().After(expiresAt)
}

func (g *QRGenerator) DisplayQRCodeInTerminal(qrCode, sessionID string) {

	qrterminal.GenerateHalfBlock(qrCode, qrterminal.L, os.Stdout)
	fmt.Printf("QR code for session %s:\n%s\n", strings.ToUpper(sessionID), qrCode)

	g.logger.InfoWithFields("QR code displayed in terminal", map[string]interface{}{
		"session_id": sessionID,
		"qr_length":  len(qrCode),
	})
}
