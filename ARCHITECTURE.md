# 🏗️ Arquitetura zpwoot - Clean Architecture Pragmática

## 📋 Visão Geral

O **zpwoot** utiliza uma **Clean Architecture Pragmática** focada em simplicidade, manutenibilidade e performance. Esta arquitetura evita over-engineering mantendo separação clara de responsabilidades.

## 🎯 Princípios Fundamentais

### 1. **Separação de Responsabilidades**
- Cada camada tem uma responsabilidade específica e bem definida
- Dependências fluem sempre para dentro (Dependency Inversion)
- Lógica de negócio isolada de detalhes de implementação

### 2. **Simplicidade Pragmática**
- Evita abstrações desnecessárias
- Foca em resolver problemas reais do negócio
- Prioriza legibilidade e manutenibilidade

### 3. **Testabilidade**
- Interfaces bem definidas para mocking
- Lógica de negócio testável isoladamente
- Testes unitários, integração e end-to-end

## 🏛️ Estrutura de Camadas

```
zpwoot/
├── internal/
│   ├── core/                    # 🎯 Core Business Logic
│   ├── services/                # 🔧 Application Services
│   └── adapters/                # 🔌 External Connections
├── platform/                   # 🏗️ Infrastructure
└── cmd/                        # 🚀 Entry Points
```

## 📁 Detalhamento das Camadas

### 🎯 **Core** - Lógica de Negócio Central

**Responsabilidade:** Contém as regras de negócio puras, entidades e contratos.

**Estrutura:**
```
internal/core/
├── session/
│   ├── models.go               # Entidades e Value Objects
│   ├── service.go              # Regras de negócio
│   └── contracts.go            # Interfaces (Repository, Gateway)
├── messaging/
├── contacts/
├── integrations/
└── shared/                     # Código compartilhado
    ├── errors/
    ├── events/
    └── types/
```

**Regras:**
- ✅ **PODE:** Definir entidades, value objects, regras de negócio
- ✅ **PODE:** Definir interfaces para dependências externas
- ✅ **PODE:** Usar outros módulos do core via interfaces
- ❌ **NÃO PODE:** Importar adapters, services ou platform
- ❌ **NÃO PODE:** Conhecer detalhes de HTTP, banco de dados, APIs externas
- ❌ **NÃO PODE:** Ter dependências de frameworks externos

**Exemplo de Import Válido:**
```go
// ✅ Permitido
import "zpwoot/internal/core/shared/errors"
import "zpwoot/internal/core/session"

// ❌ Proibido
import "zpwoot/internal/adapters/database"
import "zpwoot/internal/services"
import "github.com/gin-gonic/gin"
```

### 🔧 **Services** - Serviços de Aplicação

**Responsabilidade:** Orquestra operações, coordena entre core e adapters.

**Estrutura:**
```
internal/services/
├── session_service.go          # Orquestração de sessões
├── message_service.go          # Orquestração de mensagens
├── integration_service.go      # Orquestração de integrações
└── shared/
    ├── validation/
    └── mapping/
```

**Regras:**
- ✅ **PODE:** Importar e usar core
- ✅ **PODE:** Definir DTOs e requests/responses
- ✅ **PODE:** Coordenar múltiplas operações do core
- ✅ **PODE:** Implementar validações de entrada
- ❌ **NÃO PODE:** Conter lógica de negócio complexa
- ❌ **NÃO PODE:** Acessar adapters diretamente (usar via DI)
- ❌ **NÃO PODE:** Conhecer detalhes de implementação externa

**Exemplo de Service:**
```go
type SessionService struct {
    sessionCore *session.Service  // ✅ Usa core
    logger      Logger            // ✅ Interface injetada
}

func (s *SessionService) CreateSession(req *CreateSessionRequest) error {
    // ✅ Validação de entrada
    // ✅ Coordenação com core
    // ✅ Mapeamento de DTOs
}
```

### 🔌 **Adapters** - Conexões Externas

**Responsabilidade:** Implementa interfaces definidas no core, conecta com mundo externo.

**Estrutura:**
```
internal/adapters/
├── http/                       # REST API, handlers
│   ├── handlers/
│   ├── middleware/
│   └── routes/
├── database/                   # Implementações de Repository
│   ├── postgres/
│   └── migrations/
├── whatsapp/                   # Gateway WhatsApp
├── chatwoot/                   # Gateway Chatwoot
└── events/                     # Event Publishers
```

**Regras:**
- ✅ **PODE:** Implementar interfaces do core
- ✅ **PODE:** Usar frameworks e bibliotecas externas
- ✅ **PODE:** Conhecer detalhes de protocolos (HTTP, SQL, etc.)
- ✅ **PODE:** Importar services para injeção de dependência
- ❌ **NÃO PODE:** Conter lógica de negócio
- ❌ **NÃO PODE:** Importar outros adapters diretamente
- ❌ **NÃO PODE:** Modificar entidades do core

**Exemplo de Adapter:**
```go
// ✅ Implementa interface do core
type PostgresSessionRepository struct {
    db *sql.DB
}

func (r *PostgresSessionRepository) Save(ctx context.Context, session *session.Session) error {
    // ✅ Detalhes de persistência
    // ✅ Mapeamento entidade -> SQL
}
```

### 🏗️ **Platform** - Infraestrutura

**Responsabilidade:** Configuração, logging, monitoring, utilitários.

**Estrutura:**
```
platform/
├── config/                     # Configurações
├── logging/                    # Sistema de logs
├── monitoring/                 # Métricas e health checks
├── database/                   # Conexão e migrations
└── container/                  # Dependency Injection
```

**Regras:**
- ✅ **PODE:** Configurar e inicializar dependências
- ✅ **PODE:** Implementar cross-cutting concerns
- ✅ **PODE:** Usar qualquer biblioteca externa
- ❌ **NÃO PODE:** Conter lógica de negócio
- ❌ **NÃO PODE:** Conhecer detalhes específicos do domínio

### 🚀 **CMD** - Entry Points

**Responsabilidade:** Pontos de entrada da aplicação.

**Estrutura:**
```
cmd/
├── server/                     # HTTP Server
├── worker/                     # Background workers
└── cli/                        # CLI tools
```

**Regras:**
- ✅ **PODE:** Inicializar aplicação
- ✅ **PODE:** Configurar dependency injection
- ✅ **PODE:** Importar qualquer camada
- ❌ **NÃO PODE:** Conter lógica de negócio
- ❌ **NÃO PODE:** Conter lógica de aplicação

## 🔄 Fluxo de Dependências

```
cmd → platform → adapters → services → core
 ↑                                        ↓
 └────────── interfaces ←─────────────────┘
```

**Regra de Ouro:** Dependências sempre fluem para dentro (core não depende de nada externo).

## 📋 Regras de Import

### ✅ **Imports Permitidos**

```go
// Core pode importar
import "zpwoot/internal/core/shared"
import "zpwoot/internal/core/session"

// Services pode importar
import "zpwoot/internal/core/session"
import "zpwoot/internal/core/shared"

// Adapters pode importar
import "zpwoot/internal/core/session"
import "zpwoot/internal/services"
import "github.com/gin-gonic/gin"

// Platform pode importar
import "zpwoot/internal/adapters"
import "zpwoot/internal/services"
import "zpwoot/internal/core"

// CMD pode importar
import "zpwoot/platform"
import "zpwoot/internal/adapters"
import "zpwoot/internal/services"
import "zpwoot/internal/core"
```

### ❌ **Imports Proibidos**

```go
// Core NÃO pode importar
import "zpwoot/internal/services"        // ❌
import "zpwoot/internal/adapters"        // ❌
import "zpwoot/platform"                 // ❌
import "github.com/gin-gonic/gin"        // ❌

// Services NÃO pode importar
import "zpwoot/internal/adapters"        // ❌
import "zpwoot/platform"                 // ❌

// Adapters NÃO pode importar
import "zpwoot/platform"                 // ❌ (exceto para DI)
```

## 🧪 Estratégia de Testes

### **Testes Unitários**
- **Core:** Testa lógica de negócio isoladamente
- **Services:** Testa orquestração com mocks
- **Adapters:** Testa implementações específicas

### **Testes de Integração**
- Testa interação entre camadas
- Usa banco de dados de teste
- Testa APIs externas com mocks

### **Testes End-to-End**
- Testa fluxos completos
- Usa ambiente similar à produção

## 🔧 Dependency Injection

### **Container de DI**
```go
type Container struct {
    // Core
    SessionService *session.Service
    
    // Services
    SessionAppService *services.SessionService
    
    // Adapters
    SessionRepo session.Repository
    WhatsAppGW  session.WhatsAppGateway
}
```

### **Inicialização**
```go
func NewContainer() *Container {
    // 1. Criar adapters
    sessionRepo := postgres.NewSessionRepository(db)
    whatsappGW := whatsapp.NewGateway(client)
    
    // 2. Criar core services
    sessionCore := session.NewService(sessionRepo, whatsappGW)
    
    // 3. Criar application services
    sessionApp := services.NewSessionService(sessionCore)
    
    return &Container{...}
}
```

## 📊 Métricas de Qualidade

### **Acoplamento**
- Core: 0 dependências externas
- Services: Apenas core
- Adapters: Core + Services (via DI)

### **Coesão**
- Cada módulo tem responsabilidade única
- Funcionalidades relacionadas agrupadas

### **Testabilidade**
- 100% das interfaces mockáveis
- Lógica de negócio testável isoladamente

## 🚨 Violações Comuns

### ❌ **Anti-Patterns a Evitar**

1. **Core importando adapters**
```go
// ❌ ERRADO
import "zpwoot/adapters/database"
```

2. **Lógica de negócio em adapters**
```go
// ❌ ERRADO
func (h *HTTPHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
    if session.Status == "active" && session.LastSeen > time.Now() {
        // ❌ Lógica de negócio no adapter
    }
}
```

3. **Services acessando adapters diretamente**
```go
// ❌ ERRADO
func (s *SessionService) CreateSession() {
    db := postgres.Connect() // ❌ Acesso direto
}
```

## ✅ **Checklist de Conformidade**

- [ ] Core não importa nenhuma camada externa
- [ ] Services só importa core
- [ ] Adapters implementam interfaces do core
- [ ] Todas as dependências são injetadas
- [ ] Lógica de negócio está no core
- [ ] Testes cobrem todas as camadas
- [ ] Documentação está atualizada

## 🔍 Ferramentas de Validação

### **Linter Customizado**
```bash
# Verificar violações de arquitetura
make arch-lint

# Verificar imports proibidos
make import-check

# Verificar cobertura de testes por camada
make test-coverage
```

### **Scripts de Validação**
```bash
#!/bin/bash
# scripts/validate-architecture.sh

# Verificar se core não importa camadas externas
if grep -r "zpwoot/adapters\|zpwoot/services\|zpwoot/platform" core/; then
    echo "❌ ERRO: Core não pode importar camadas externas"
    exit 1
fi

echo "✅ Arquitetura válida"
```

## 📚 Exemplos Práticos

### **Adicionando Nova Feature**

1. **Criar módulo no core:**
```go
// core/newsletter/models.go
type Newsletter struct {
    ID      string
    Title   string
    Content string
}

// core/newsletter/service.go
type Service struct {
    repo Repository
}

// core/newsletter/contracts.go
type Repository interface {
    Save(ctx context.Context, newsletter *Newsletter) error
}
```

2. **Criar service de aplicação:**
```go
// services/newsletter_service.go
type NewsletterService struct {
    newsletterCore *newsletter.Service
}
```

3. **Implementar adapters:**
```go
// adapters/database/newsletter_repository.go
type PostgresNewsletterRepository struct {
    db *sql.DB
}

// adapters/http/newsletter_handler.go
type NewsletterHandler struct {
    service *services.NewsletterService
}
```

### **Migrando Código Legacy**

1. **Identificar lógica de negócio**
2. **Extrair para core**
3. **Criar interfaces**
4. **Implementar adapters**
5. **Criar services**
6. **Atualizar DI**

## 🎓 Treinamento da Equipe

### **Onboarding Checklist**
- [ ] Ler este documento
- [ ] Entender fluxo de dependências
- [ ] Praticar com feature simples
- [ ] Code review com arquiteto
- [ ] Validar com ferramentas

### **Code Review Guidelines**
- Verificar imports
- Validar responsabilidades
- Checar testes
- Confirmar interfaces

---

**Esta arquitetura garante:**
- 🎯 **Foco no negócio** - lógica isolada e testável
- 🔧 **Flexibilidade** - fácil troca de implementações
- 🚀 **Performance** - sem overhead desnecessário
- 📈 **Escalabilidade** - crescimento organizado
- 🛡️ **Manutenibilidade** - código limpo e estruturado

**Versão:** 1.0
**Última atualização:** 2025-01-01
**Responsável:** Equipe zpwoot
