# 🚀 Plano de Migração Legacy → Clean Architecture

## 📋 Visão Geral

Este documento detalha o plano completo para migrar o código legacy do zpwoot para a nova arquitetura Clean Architecture definida em `ARCHITECTURE.md`. A migração será realizada de forma **incremental e segura**, mantendo a aplicação funcional durante todo o processo.

## 🎯 Objetivos da Migração

### **Principais Metas**
- ✅ **Conformidade arquitetural**: Seguir rigorosamente as regras da Clean Architecture
- ✅ **Zero downtime**: Manter aplicação funcional durante migração
- ✅ **Melhoria da testabilidade**: Implementar testes abrangentes
- ✅ **Redução do acoplamento**: Separar responsabilidades claramente
- ✅ **Facilitar manutenção**: Código mais limpo e organizados

### **Benefícios Esperados**
- 🎯 **Lógica de negócio isolada** e testável
- 🔧 **Flexibilidade** para trocar implementações
- 📈 **Escalabilidade** organizada
- 🛡️ **Manutenibilidade** aprimorada
- 🧪 **Cobertura de testes** completa

## 📊 Análise da Estrutura Legacy

### **Mapeamento Atual → Destino**

| **Legacy** | **Nova Arquitetura** | **Responsabilidade** |
|------------|---------------------|---------------------|
| `legacy/internal/domain/*` | `internal/core/*` | Lógica de negócio pura |
| `legacy/internal/app/*` | `internal/services/*` | Orquestração e coordenação |
| `legacy/internal/infra/*` | `internal/adapters/*` | Conexões externas |
| `legacy/internal/ports/*` | `internal/core/*/contracts.go` | Interfaces e contratos |
| `platform/*` | `platform/*` | Infraestrutura (mantém) |
| `cmd/*` | `cmd/*` | Entry points (mantém) |

### **Módulos Identificados**

#### **1. Core Business Logic** (legacy/internal/domain → internal/core/)
- **Session**: Gerenciamento de sessões WhatsApp
- **Message**: Lógica de mensagens e validações
- **Contact**: Regras de contatos e validações
- **Group**: Lógica de grupos e participantes
- **Media**: Regras de mídia e cache
- **Chatwoot**: Lógica de integração
- **Webhook**: Regras de notificações
- **Newsletter**: Lógica de newsletters
- **Community**: Regras de comunidades

#### **2. Application Services** (legacy/internal/app → internal/services/)
- **Session Service**: Orquestração de sessões
- **Message Service**: Coordenação de envio/recebimento
- **Contact Service**: Orquestração de contatos
- **Group Service**: Coordenação de grupos
- **Media Service**: Orquestração de mídia
- **Chatwoot Service**: Coordenação de integração
- **Webhook Service**: Orquestração de webhooks
- **Newsletter Service**: Coordenação de newsletters
- **Community Service**: Coordenação de comunidades

#### **3. External Adapters** (legacy/internal/infra → internal/adapters/)
- **HTTP Adapters**: REST API handlers
- **Database Adapters**: Implementações de Repository
- **WhatsApp Adapter**: Gateway WhatsApp (wameow)
- **Chatwoot Adapter**: Gateway Chatwoot
- **Webhook Adapter**: Event publishers

## 🗺️ Estratégia de Migração

### **Abordagem: Strangler Fig Pattern**

Utilizaremos o padrão **Strangler Fig** para migração incremental:

1. **Criar nova estrutura** em paralelo ao legacy
2. **Migrar módulo por módulo** mantendo compatibilidade
3. **Redirecionar tráfego** gradualmente para nova implementação
4. **Remover código legacy** após validação completa

### **Fases da Migração**

#### **📋 Fase 1: Preparação e Fundação (Semana 1-2)**
- [ ] Criar estrutura de diretórios da nova arquitetura
- [ ] Implementar ferramentas de validação arquitetural
- [ ] Configurar testes e CI/CD para nova estrutura
- [ ] Documentar padrões e convenções

#### **🎯 Fase 2: Core Business Logic (Semana 3-6)**
- [ ] Migrar entidades e value objects
- [ ] Extrair lógica de negócio pura dos domain services
- [ ] Definir interfaces (contracts) no core
- [ ] Implementar testes unitários para core

#### **🔧 Fase 3: Application Services (Semana 7-10)**
- [ ] Criar services de aplicação
- [ ] Implementar orquestração entre core e adapters
- [ ] Migrar use cases do legacy/internal/app
- [ ] Implementar testes de integração

#### **🔌 Fase 4: External Adapters (Semana 11-14)**
- [ ] Implementar adapters HTTP
- [ ] Migrar repositories para adapters/database
- [ ] Implementar gateways externos
- [ ] Configurar dependency injection

#### **🚀 Fase 5: Integration & Cleanup (Semana 15-16)**
- [ ] Integrar todas as camadas
- [ ] Executar testes end-to-end
- [ ] Remover código legacy
- [ ] Documentação final

## 📁 Estrutura de Diretórios Detalhada

### **Nova Estrutura Completa**
```
zpwoot/
├── internal/                       # 📦 Internal packages
│   ├── core/                      # 🎯 Core Business Logic
│   │   ├── session/
│   │   │   ├── models.go          # Entidades Session, DeviceInfo, etc.
│   │   │   ├── service.go         # Regras de negócio de sessão
│   │   │   ├── contracts.go       # Repository, WameowGateway interfaces
│   │   │   └── errors.go          # Erros específicos do domínio
│   │   ├── messaging/
│   │   │   ├── models.go          # Message, MessageType, etc.
│   │   │   ├── service.go         # Validações e regras de mensagem
│   │   │   ├── contracts.go       # MessageRepository, WhatsAppGateway
│   │   │   └── errors.go
│   │   ├── contacts/
│   │   │   ├── models.go          # Contact, ContactInfo, etc.
│   │   │   ├── service.go         # Validações de contato
│   │   │   ├── contracts.go       # ContactRepository interface
│   │   │   └── errors.go
│   │   ├── groups/
│   │   │   ├── models.go          # Group, Participant, etc.
│   │   │   ├── service.go         # Regras de grupo
│   │   │   ├── contracts.go       # GroupRepository interface
│   │   │   └── errors.go
│   │   ├── integrations/
│   │   │   ├── chatwoot/
│   │   │   │   ├── models.go      # ChatwootConfig, etc.
│   │   │   │   ├── service.go     # Lógica de integração
│   │   │   │   ├── contracts.go   # ChatwootGateway interface
│   │   │   │   └── errors.go
│   │   │   └── webhook/
│   │   │       ├── models.go      # WebhookConfig, Event, etc.
│   │   │       ├── service.go     # Regras de webhook
│   │   │       ├── contracts.go   # WebhookGateway interface
│   │   │       └── errors.go
│   │   └── shared/
│   │       ├── errors/
│   │       │   ├── domain.go      # Erros de domínio base
│   │       │   └── codes.go       # Códigos de erro
│   │       ├── events/
│   │       │   ├── event.go       # Event base
│   │       │   └── publisher.go   # EventPublisher interface
│   │       └── types/
│   │           ├── id.go          # ID types
│   │           ├── time.go        # Time utilities
│   │           └── validation.go  # Validation helpers
│   ├── services/                  # 🔧 Application Services
│   │   ├── session_service.go     # Orquestração de sessões
│   │   ├── message_service.go     # Orquestração de mensagens
│   │   ├── contact_service.go     # Orquestração de contatos
│   │   ├── group_service.go       # Orquestração de grupos
│   │   ├── chatwoot_service.go    # Orquestração Chatwoot
│   │   ├── webhook_service.go     # Orquestração webhooks
│   │   └── shared/
│   │       ├── validation/        # Validações de entrada
│   │       └── mapping/           # DTOs e mapeamentos
│   └── adapters/                  # 🔌 External Connections
│       ├── http/
│       │   ├── handlers/
│       │   │   ├── session_handler.go
│       │   │   ├── message_handler.go
│       │   │   ├── contact_handler.go
│       │   │   ├── group_handler.go
│       │   │   ├── chatwoot_handler.go
│       │   │   └── webhook_handler.go
│       │   ├── middleware/
│       │   │   ├── auth.go
│       │   │   ├── cors.go
│       │   │   ├── logging.go
│       │   │   └── validation.go
│       │   └── routes/
│       │       ├── routes.go
│       │       └── swagger.go
│       ├── database/
│       │   ├── postgres/
│       │   │   ├── session_repository.go
│       │   │   ├── message_repository.go
│       │   │   ├── contact_repository.go
│       │   │   ├── group_repository.go
│       │   │   ├── chatwoot_repository.go
│       │   │   └── webhook_repository.go
│       │   └── migrations/
│       │       ├── 001_initial.up.sql
│       │       └── 001_initial.down.sql
│       ├── whatsapp/
│       │   ├── gateway.go         # WhatsApp Gateway implementation
│       │   ├── client.go          # WhatsApp client wrapper
│       │   ├── events.go          # Event handling
│       │   └── mapper.go          # Data mapping
│       ├── chatwoot/
│       │   ├── gateway.go         # Chatwoot Gateway implementation
│       │   ├── client.go          # HTTP client
│       │   ├── webhook.go         # Webhook handling
│       │   └── mapper.go          # Data mapping
│       └── events/
│           ├── publisher.go       # Event publisher implementation
│           └── handlers.go        # Event handlers
├── platform/                      # 🏗️ Infrastructure (mantém atual)
├── cmd/                           # 🚀 Entry Points (mantém atual)
└── legacy/                        # 📦 Código legacy (temporário)
```

## 🔄 Plano de Migração Detalhado

### **Módulo 1: Session (Prioridade Alta)**

#### **Semana 3: Core Session**
- [ ] Criar `internal/core/session/models.go`
  - Migrar `Session`, `DeviceInfo`, `ProxyConfig` de `legacy/internal/domain/session/entity.go`
  - Limpar dependências externas
- [ ] Criar `internal/core/session/service.go`
  - Extrair lógica pura de `legacy/internal/domain/session/service.go`
  - Remover dependências de infraestrutura
- [ ] Criar `internal/core/session/contracts.go`
  - Definir `Repository`, `WameowGateway`, `QRGenerator` interfaces
- [ ] Implementar testes unitários

#### **Semana 4: Session Service**
- [ ] Criar `internal/services/session_service.go`
  - Migrar orquestração de `legacy/internal/app/session/usecase.go`
  - Implementar validações de entrada
- [ ] Implementar DTOs e mapeamentos
- [ ] Testes de integração

#### **Semana 5: Session Adapters**
- [ ] Criar `internal/adapters/database/postgres/session_repository.go`
  - Migrar de `legacy/internal/infra/repository/session_repository.go`
- [ ] Criar `internal/adapters/whatsapp/gateway.go` (parte session)
  - Migrar de `legacy/internal/infra/wameow/manager.go`
- [ ] Criar `internal/adapters/http/handlers/session_handler.go`
  - Migrar handlers HTTP

### **Módulo 2: Messaging (Prioridade Alta)**

#### **Semana 6: Core Messaging**
- [ ] Criar estrutura internal/core/messaging
- [ ] Migrar entidades de mensagem
- [ ] Extrair lógica de validação
- [ ] Testes unitários

#### **Semana 7: Messaging Service & Adapters**
- [ ] Implementar internal/services/message_service.go
- [ ] Migrar adapters de mensagem
- [ ] Integrar com WhatsApp gateway

### **Módulos 3-6: Contacts, Groups, Chatwoot, Webhooks**

#### **Semanas 8-12: Migração Paralela**
- [ ] Seguir mesmo padrão: Core → Services → Adapters
- [ ] Manter compatibilidade com legacy
- [ ] Implementar testes progressivamente

### **Fase Final: Integration & Cleanup**

#### **Semanas 13-14: Integração**
- [ ] Configurar dependency injection completo
- [ ] Implementar roteamento híbrido (legacy + novo)
- [ ] Testes end-to-end completos

#### **Semanas 15-16: Cleanup**
- [ ] Remover código legacy gradualmente
- [ ] Atualizar documentação
- [ ] Validação final de arquitetura

## 🧪 Estratégia de Testes

### **Pirâmide de Testes**
```
        /\
       /  \
      / E2E \     ← Poucos, mas críticos
     /______\
    /        \
   / Integration \  ← Testes de integração
  /______________\
 /                \
/ Unit Tests       \  ← Muitos, rápidos, isolados
\__________________/
```

### **Cobertura por Camada**
- **Core**: 100% cobertura unitária (lógica de negócio)
- **Services**: 90% cobertura (orquestração + mocks)
- **Adapters**: 80% cobertura (implementações específicas)

### **Ferramentas de Teste**
- **Unit**: Go testing + testify
- **Integration**: Testcontainers para DB
- **E2E**: Testes de API completos
- **Mocks**: Mockery para interfaces

## 🔧 Ferramentas de Validação

### **Linter Arquitetural**
```bash
# Verificar violações de imports
make arch-lint

# Verificar dependências proibidas
make import-check

# Validar estrutura de diretórios
make structure-check
```

### **Scripts de Validação**
- `scripts/validate-architecture.sh`: Verificar regras de import
- `scripts/check-dependencies.sh`: Validar fluxo de dependências
- `scripts/test-coverage.sh`: Verificar cobertura por camada

## 📈 Métricas de Sucesso

### **Indicadores de Qualidade**
- [ ] **0 violações** de regras arquiteturais
- [ ] **90%+ cobertura** de testes
- [ ] **0 dependências circulares**
- [ ] **100% interfaces** mockáveis
- [ ] **Documentação** atualizada

### **Performance**
- [ ] **Tempo de build** ≤ atual
- [ ] **Tempo de testes** ≤ 2x atual
- [ ] **Memory footprint** ≤ atual
- [ ] **API response time** ≤ atual

## ⚠️ Riscos e Mitigações

### **Riscos Identificados**
1. **Quebra de funcionalidade** durante migração
   - **Mitigação**: Testes abrangentes + migração incremental
2. **Aumento de complexidade** temporária
   - **Mitigação**: Documentação clara + code reviews
3. **Resistência da equipe** a mudanças
   - **Mitigação**: Treinamento + benefícios claros

### **Plano de Rollback**
- Manter legacy funcional até validação completa
- Feature flags para alternar implementações
- Monitoramento contínuo de métricas

## 📅 Cronograma Resumido

| **Fase** | **Duração** | **Entregáveis** |
|----------|-------------|-----------------|
| Preparação | 2 semanas | Estrutura + ferramentas |
| Core Migration | 4 semanas | Lógica de negócio isolada |
| Services Migration | 4 semanas | Orquestração implementada |
| Adapters Migration | 4 semanas | Conexões externas |
| Integration & Cleanup | 2 semanas | Sistema completo |

**Total: 16 semanas (~4 meses)**

## 🎯 Próximos Passos

1. **Aprovação do plano** pela equipe
2. **Setup do ambiente** de desenvolvimento
3. **Início da Fase 1**: Preparação e fundação
4. **Definição de responsáveis** por módulo
5. **Configuração de ferramentas** de validação

---

**Este plano garante uma migração segura, incremental e bem estruturada para a nova arquitetura Clean Architecture, mantendo a qualidade e funcionalidade do zpwoot.**
