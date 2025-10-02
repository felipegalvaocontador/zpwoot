# wameow - Wameow Integration Module

## 📋 Visão Geral

O módulo `wameow` implementa a integração completa com Wameow Web usando a biblioteca `whatsmeow`. Ele fornece uma abstração robusta para gerenciar sessões Wameow, conexões, eventos e QR codes.

## 🏗️ Arquitetura

```
internal/infra/wameow/
├── README.md           # Este arquivo
├── manager.go          # Manager principal - implementa WameowManager
├── connection.go       # Gerenciamento de conexões
├── events.go           # Manipulação de eventos Wameow
├── utils.go            # Utilitários e validações
└── config.go           # Configuração e factory
```

## 🎯 Componentes Principais

### **Manager**
- **Arquivo**: `manager.go`
- **Responsabilidade**: Implementa a interface `WameowManager`
- **Funcionalidades**:
  - Criação e gerenciamento de sessões
  - Conexão/desconexão de clientes
  - Geração de QR codes
  - Configuração de proxy
  - Informações de dispositivo

### **ConnectionManager**
- **Arquivo**: `connection.go`
- **Responsabilidade**: Gerencia conexões Wameow
- **Funcionalidades**:
  - Conexão segura com retry
  - Desconexão segura
  - Validação de clientes
  - Tratamento de erros

### **QRCodeGenerator**
- **Arquivo**: `connection.go`
- **Responsabilidade**: Geração e exibição de QR codes
- **Funcionalidades**:
  - Geração de imagem base64
  - Exibição no terminal
  - Tratamento de erros

### **SessionManager**
- **Arquivo**: `connection.go`
- **Responsabilidade**: Gerencia estado das sessões
- **Funcionalidades**:
  - Atualização de status
  - Atualização de QR codes
  - Sincronização com banco de dados

### **EventHandler**
- **Arquivo**: `events.go`
- **Responsabilidade**: Manipula eventos Wameow
- **Eventos Suportados**:
  - Conexão/desconexão
  - QR code
  - Pareamento
  - Mensagens
  - Recibos
  - Presença

## 🚀 Como Usar

### **1. Criação do Manager**

```go
import (
    "zpwoot/internal/infra/wameow"
    "zpwoot/platform/logger"
)

// Usando o builder pattern
manager, err := wameow.NewManagerBuilder().
    WithLogger(logger).
    WithSessionRepository(sessionRepo).
    WithDatabase(db).
    WithConfig(config).
    Build()

if err != nil {
    log.Fatal("Failed to create Wameow manager:", err)
}
```

### **2. Criação de Sessão**

```go
// Criar nova sessão
err := manager.CreateSession("session-123", nil)
if err != nil {
    log.Error("Failed to create session:", err)
}

// Conectar sessão
err = manager.ConnectSession("session-123")
if err != nil {
    log.Error("Failed to connect session:", err)
}
```

### **3. Configuração com Proxy**

```go
proxyConfig := &session.ProxyConfig{
    Type:     "http",
    Host:     "proxy.example.com",
    Port:     8080,
    Username: "user",
    Password: "pass",
}

err := manager.CreateSession("session-123", proxyConfig)
```

### **4. Obter QR Code**

```go
qrResponse, err := manager.GetQRCode("session-123")
if err != nil {
    log.Error("Failed to get QR code:", err)
    return
}

fmt.Printf("QR Code: %s\n", qrResponse.QRCode)
fmt.Printf("Expires at: %s\n", qrResponse.ExpiresAt)
```

## ⚙️ Configuração

### **Config Struct**

```go
type Config struct {
    DatabaseURL      string        `json:"database_url"`
    LogLevel         string        `json:"log_level"`
    SessionTimeout   time.Duration `json:"session_timeout"`
    RetryAttempts    int           `json:"retry_attempts"`
    RetryInterval    time.Duration `json:"retry_interval"`
    QRCodeTimeout    time.Duration `json:"qr_code_timeout"`
    EnableQRTerminal bool          `json:"enable_qr_terminal"`
}
```

### **Configuração Padrão**

```go
config := wameow.DefaultConfig()
// Personalizar conforme necessário
config.RetryAttempts = 5
config.SessionTimeout = 60 * time.Minute
```

## 📊 Eventos Wameow

### **Eventos Suportados**

| Evento | Descrição | Ação |
|--------|-----------|------|
| `Connected` | Cliente conectado | Atualiza status para `connected` |
| `Disconnected` | Cliente desconectado | Atualiza status para `disconnected` |
| `QR` | QR code recebido | Gera imagem e exibe no terminal |
| `PairSuccess` | Pareamento bem-sucedido | Atualiza device JID |
| `PairError` | Erro no pareamento | Atualiza status para `error` |
| `Message` | Mensagem recebida | Processa e encaminha |
| `Receipt` | Recibo de mensagem | Log e processamento |
| `LoggedOut` | Logout realizado | Limpa sessão |

### **Customização de Eventos**

```go
// Os eventos são automaticamente manipulados
// Você pode estender a funcionalidade modificando events.go
```

## 🔧 Utilitários

### **Validações**

```go
// Validar cliente
err := wameow.ValidateClientAndStore(client, sessionID)

// Validar JID
isValid := wameow.IsValidJID("5511999999999@s.Wameow.net")

// Validar session ID
err := wameow.ValidateSessionID("session-123")
```

### **Informações**

```go
// Status da conexão
status := wameow.GetConnectionStatus(client, sessionID)

// Informações do cliente
info := wameow.GetClientInfo(client)

// Health check
health := manager.HealthCheck()
```

## 🛡️ Tratamento de Erros

### **Tipos de Erro**

```go
// ConnectionError - erros de conexão
type ConnectionError struct {
    SessionID string
    Operation string
    Err       error
}

// Verificar tipo de erro
if connErr, ok := err.(*wameow.ConnectionError); ok {
    log.Printf("Connection error for session %s: %v", 
               connErr.SessionID, connErr.Err)
}
```

### **Categorias de Erro**

- `connection` - Problemas de conectividade
- `authentication` - Problemas de autenticação
- `timeout` - Timeouts
- `network` - Problemas de rede
- `unknown` - Erros não categorizados

## 📈 Monitoramento

### **Health Check**

```go
health := manager.HealthCheck()
fmt.Printf("Total sessions: %d\n", health["total_sessions"])
fmt.Printf("Connected: %d\n", health["connected_sessions"])
fmt.Printf("Logged in: %d\n", health["logged_in_sessions"])
```

### **Estatísticas**

```go
stats := manager.GetStats()
// Retorna as mesmas informações do health check
```

## 🔄 Retry e Reconexão

### **Configuração de Retry**

```go
retryConfig := &wameow.RetryConfig{
    MaxRetries:    5,
    RetryInterval: 30 * time.Second,
}

err := manager.ConnectWithRetry(client, sessionID, retryConfig)
```

### **Reconexão Automática**

O módulo implementa reconexão automática através dos event handlers. Quando uma desconexão é detectada, o sistema pode tentar reconectar automaticamente.

## 🧪 Testes

### **Testes Unitários**

```bash
# Executar testes do módulo
go test ./internal/infra/wameow/...

# Com cobertura
go test -cover ./internal/infra/wameow/...
```

### **Testes de Integração**

```bash
# Testes com banco real
go test -tags=integration ./internal/infra/wameow/...
```

## 📚 Dependências

### **Principais**

- `go.mau.fi/whatsmeow` - Cliente Wameow Web
- `github.com/skip2/go-qrcode` - Geração de QR codes
- `github.com/mdp/qrterminal/v3` - QR codes no terminal

### **Internas**

- `zpwoot/internal/domain/session` - Entidades de sessão
- `zpwoot/internal/ports` - Interfaces
- `zpwoot/platform/logger` - Sistema de logging

## 🚨 Limitações Conhecidas

1. **Phone Pairing**: Não implementado ainda
2. **Proxy Support**: Configuração básica, sem validação completa
3. **Message Sending**: Não implementado neste módulo
4. **Media Handling**: Não implementado

## 🔮 Próximos Passos

1. **Implementar envio de mensagens**
2. **Adicionar suporte completo a proxy**
3. **Implementar phone pairing**
4. **Adicionar testes abrangentes**
5. **Melhorar tratamento de erros**
6. **Adicionar métricas detalhadas**

---

**Status**: ✅ Implementação base completa
**Última atualização**: 2024-01-01
