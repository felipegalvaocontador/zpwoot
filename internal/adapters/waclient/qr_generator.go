package waclient

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image/png"
	"os"
	"strings"
	"time"

	"github.com/mdp/qrterminal/v3"
	"github.com/skip2/go-qrcode"

	"zpwoot/internal/core/session"
	"zpwoot/platform/logger"
)

// QRGenerator implementa session.QRCodeGenerator
type QRGenerator struct {
	logger *logger.Logger
}

// NewQRGenerator cria nova instância do gerador de QR code
func NewQRGenerator(logger *logger.Logger) *QRGenerator {
	return &QRGenerator{
		logger: logger,
	}
}

// Generate implementa session.QRCodeGenerator.Generate
func (g *QRGenerator) Generate(ctx context.Context, sessionName string) (*session.QRCodeResponse, error) {
	// Esta implementação será chamada pelo gateway
	// Por enquanto retorna erro pois o QR code vem dos eventos do whatsmeow
	return nil, fmt.Errorf("QR code generation is handled by WhatsApp events")
}

// GenerateQRCode gera QR code como string (método auxiliar)
func (g *QRGenerator) GenerateQRCode(data string) (string, error) {
	// Para WhatsApp, o data já é o QR code string
	// Apenas retornamos como está
	return data, nil
}

// GenerateQRCodeImage gera QR code como imagem base64
func (g *QRGenerator) GenerateQRCodeImage(data string) (string, error) {
	g.logger.DebugWithFields("Generating QR code image", map[string]interface{}{
		"data_length": len(data),
	})

	// Gerar QR code como imagem PNG
	qr, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		return "", fmt.Errorf("failed to create QR code: %w", err)
	}

	// Configurar tamanho da imagem
	qr.DisableBorder = false

	// Gerar imagem PNG
	img := qr.Image(256)

	// Converter para bytes
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("failed to encode QR code image: %w", err)
	}

	// Converter para base64
	base64Image := base64.StdEncoding.EncodeToString(buf.Bytes())
	dataURI := fmt.Sprintf("data:image/png;base64,%s", base64Image)

	g.logger.DebugWithFields("QR code image generated", map[string]interface{}{
		"image_size": len(base64Image),
	})

	return dataURI, nil
}

// GenerateQRCodePNG gera QR code como bytes PNG
func (g *QRGenerator) GenerateQRCodePNG(data string, size int) ([]byte, error) {
	if size <= 0 {
		size = 256 // tamanho padrão
	}

	g.logger.DebugWithFields("Generating QR code PNG", map[string]interface{}{
		"data_length": len(data),
		"size":        size,
	})

	// Gerar QR code
	qr, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("failed to create QR code: %w", err)
	}

	// Configurar
	qr.DisableBorder = false

	// Gerar como PNG bytes
	pngBytes, err := qr.PNG(size)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code PNG: %w", err)
	}

	g.logger.DebugWithFields("QR code PNG generated", map[string]interface{}{
		"bytes_size": len(pngBytes),
	})

	return pngBytes, nil
}

// ValidateQRCode valida se string é um QR code válido do WhatsApp
func (g *QRGenerator) ValidateQRCode(data string) bool {
	// QR codes do WhatsApp geralmente começam com números seguidos de @
	// Exemplo: "2@abc123def456..."
	if len(data) < 10 {
		return false
	}

	// Verificar se começa com dígito seguido de @
	if data[0] < '0' || data[0] > '9' {
		return false
	}

	// Procurar pelo @
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

	// Verificar se há conteúdo após o @
	if atIndex >= len(data)-1 {
		return false
	}

	return true
}

// GetQRCodeInfo extrai informações do QR code do WhatsApp
func (g *QRGenerator) GetQRCodeInfo(data string) map[string]interface{} {
	info := map[string]interface{}{
		"valid":  g.ValidateQRCode(data),
		"length": len(data),
	}

	if !g.ValidateQRCode(data) {
		return info
	}

	// Extrair versão (número antes do @)
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

// GenerateImage implementa session.QRCodeGenerator.GenerateImage
func (g *QRGenerator) GenerateImage(ctx context.Context, qrCode string) ([]byte, error) {
	return g.GenerateQRCodePNG(qrCode, 256)
}

// IsExpired implementa session.QRCodeGenerator.IsExpired
func (g *QRGenerator) IsExpired(expiresAt time.Time) bool {
	return time.Now().After(expiresAt)
}

// DisplayQRCodeInTerminal exibe QR code no terminal de forma visível
// Compatível com a interface do wameow
func (g *QRGenerator) DisplayQRCodeInTerminal(qrCode, sessionID string) {
	// Separador visual para destacar o QR code
	separator := strings.Repeat("=", 80)

	fmt.Printf("\n%s\n", separator)
	fmt.Printf("🔗 QR CODE PARA SESSÃO: %s\n", strings.ToUpper(sessionID))
	fmt.Printf("%s\n", separator)
	fmt.Println("📱 Escaneie o QR code abaixo com seu WhatsApp:")
	fmt.Println("   1. Abra o WhatsApp no seu celular")
	fmt.Println("   2. Vá em Configurações > Aparelhos conectados")
	fmt.Println("   3. Toque em 'Conectar um aparelho'")
	fmt.Println("   4. Escaneie o código QR abaixo")
	fmt.Printf("%s\n\n", separator)

	// Configurar QR terminal para melhor visualização
	config := qrterminal.Config{
		Level:     qrterminal.M,
		Writer:    os.Stdout,
		BlackChar: qrterminal.BLACK,
		WhiteChar: qrterminal.WHITE,
		QuietZone: 1,
	}

	// Exibir QR code no terminal
	qrterminal.GenerateWithConfig(qrCode, config)

	fmt.Printf("\n%s\n", separator)
	fmt.Printf("⏳ Aguardando escaneamento do QR code para sessão '%s'...\n", sessionID)
	fmt.Printf("%s\n\n", separator)

	// Log para arquivo também
	g.logger.InfoWithFields("QR code displayed in terminal", map[string]interface{}{
		"session_id": sessionID,
		"qr_length":  len(qrCode),
	})
}
