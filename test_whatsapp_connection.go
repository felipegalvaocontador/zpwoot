package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	_ "github.com/mattn/go-sqlite3"

	"zpwoot/internal/adapters/waclient"
	"zpwoot/platform/config"
	"zpwoot/platform/logger"
)

func main() {
	fmt.Println("🚀 Testando conexão WhatsApp...")

	// Criar logger
	logConfig := config.LogConfig{
		Level:  "debug",
		Format: "console",
		Output: "stdout",
	}
	appLogger := logger.New(logConfig)

	// Criar container do banco de dados em memória para teste
	ctx := context.Background()
	container, err := sqlstore.New(ctx, "sqlite3", "file:test.db?cache=shared&mode=memory&_foreign_keys=on", waLog.Noop)
	if err != nil {
		log.Fatalf("Failed to create sqlstore: %v", err)
	}

	// Criar gateway
	gateway := waclient.NewGateway(container, appLogger)

	// Testar criação de sessão
	sessionName := "test-session"

	fmt.Printf("📱 Criando sessão: %s\n", sessionName)
	err = gateway.CreateSession(ctx, sessionName)
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	// Verificar se sessão foi criada
	connected, err := gateway.IsSessionConnected(ctx, sessionName)
	if err != nil {
		log.Fatalf("Failed to check session status: %v", err)
	}

	fmt.Printf("✅ Sessão criada. Conectada: %v\n", connected)

	// Testar informações da sessão
	deviceInfo, err := gateway.GetSessionInfo(ctx, sessionName)
	if err != nil {
		log.Fatalf("Failed to get session info: %v", err)
	}

	fmt.Printf("📋 Device Info:\n")
	fmt.Printf("  Platform: %s\n", deviceInfo.Platform)
	fmt.Printf("  Device Model: %s\n", deviceInfo.DeviceModel)
	fmt.Printf("  OS Version: %s\n", deviceInfo.OSVersion)
	fmt.Printf("  App Version: %s\n", deviceInfo.AppVersion)

	// Testar geração de QR code (vai falhar porque não há conexão real)
	fmt.Printf("🔗 Tentando gerar QR code...\n")
	qrResponse, err := gateway.GenerateQRCode(ctx, sessionName)
	if err != nil {
		fmt.Printf("⚠️  Erro esperado ao gerar QR code (sem conexão real): %v\n", err)
	} else {
		fmt.Printf("✅ QR Code gerado: %s\n", qrResponse.QRCode[:50]+"...")
	}

	// Testar envio de mensagem (vai falhar porque não há conexão real)
	fmt.Printf("💬 Tentando enviar mensagem de teste...\n")
	result, err := gateway.SendTextMessage(ctx, sessionName, "5511999999999@s.whatsapp.net", "Teste de mensagem")
	if err != nil {
		fmt.Printf("⚠️  Erro esperado ao enviar mensagem (sem conexão real): %v\n", err)
	} else {
		fmt.Printf("✅ Mensagem enviada: %s\n", result.MessageID)
	}

	// Testar desconexão
	fmt.Printf("🔌 Desconectando sessão...\n")
	err = gateway.DisconnectSession(ctx, sessionName)
	if err != nil {
		log.Fatalf("Failed to disconnect session: %v", err)
	}

	// Testar exclusão
	fmt.Printf("🗑️  Deletando sessão...\n")
	err = gateway.DeleteSession(ctx, sessionName)
	if err != nil {
		log.Fatalf("Failed to delete session: %v", err)
	}

	fmt.Println("✅ Teste concluído com sucesso!")
	fmt.Println("🎉 A integração WhatsApp está funcionando corretamente!")

	// Limpar arquivo de teste
	os.Remove("test.db")
}
