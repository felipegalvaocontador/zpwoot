#!/bin/bash

# 🏗️ Script para criar estrutura da Nova Arquitetura Clean Architecture
# Executa: chmod +x scripts/create-architecture.sh && ./scripts/create-architecture.sh

set -e

echo "🚀 Criando estrutura da Nova Arquitetura Clean Architecture..."

# ==============================================
# 🎯 CORE - Business Logic
# ==============================================
echo "📁 Criando estrutura CORE..."

# Core Session
mkdir -p internal/core/session
touch internal/core/session/models.go
touch internal/core/session/service.go
touch internal/core/session/contracts.go
touch internal/core/session/errors.go

# Core Messaging
mkdir -p internal/core/messaging
touch internal/core/messaging/models.go
touch internal/core/messaging/service.go
touch internal/core/messaging/contracts.go
touch internal/core/messaging/errors.go

# Core Contacts
mkdir -p internal/core/contacts
touch internal/core/contacts/models.go
touch internal/core/contacts/service.go
touch internal/core/contacts/contracts.go
touch internal/core/contacts/errors.go

# Core Groups
mkdir -p internal/core/groups
touch internal/core/groups/models.go
touch internal/core/groups/service.go
touch internal/core/groups/contracts.go
touch internal/core/groups/errors.go

# Core Integrations - Chatwoot
mkdir -p internal/core/integrations/chatwoot
touch internal/core/integrations/chatwoot/models.go
touch internal/core/integrations/chatwoot/service.go
touch internal/core/integrations/chatwoot/contracts.go
touch internal/core/integrations/chatwoot/errors.go

# Core Integrations - Webhook
mkdir -p internal/core/integrations/webhook
touch internal/core/integrations/webhook/models.go
touch internal/core/integrations/webhook/service.go
touch internal/core/integrations/webhook/contracts.go
touch internal/core/integrations/webhook/errors.go

# Core Media
mkdir -p internal/core/media
touch internal/core/media/models.go
touch internal/core/media/service.go
touch internal/core/media/contracts.go
touch internal/core/media/errors.go

# Core Newsletter
mkdir -p internal/core/newsletter
touch internal/core/newsletter/models.go
touch internal/core/newsletter/service.go
touch internal/core/newsletter/contracts.go
touch internal/core/newsletter/errors.go

# Core Community
mkdir -p internal/core/community
touch internal/core/community/models.go
touch internal/core/community/service.go
touch internal/core/community/contracts.go
touch internal/core/community/errors.go

# Core Shared
mkdir -p internal/core/shared/errors
touch internal/core/shared/errors/domain.go
touch internal/core/shared/errors/codes.go

mkdir -p internal/core/shared/events
touch internal/core/shared/events/event.go
touch internal/core/shared/events/publisher.go

mkdir -p internal/core/shared/types
touch internal/core/shared/types/id.go
touch internal/core/shared/types/time.go
touch internal/core/shared/types/validation.go

# ==============================================
# 🔧 SERVICES - Application Services
# ==============================================
echo "📁 Criando estrutura SERVICES..."

mkdir -p internal/services
touch internal/services/session_service.go
touch internal/services/message_service.go
touch internal/services/contact_service.go
touch internal/services/group_service.go
touch internal/services/chatwoot_service.go
touch internal/services/webhook_service.go
touch internal/services/media_service.go
touch internal/services/newsletter_service.go
touch internal/services/community_service.go

# Services Shared
mkdir -p internal/services/shared/validation
touch internal/services/shared/validation/validator.go
touch internal/services/shared/validation/rules.go

mkdir -p internal/services/shared/mapping
touch internal/services/shared/mapping/dto.go
touch internal/services/shared/mapping/mapper.go

# ==============================================
# 🔌 ADAPTERS - External Connections
# ==============================================
echo "📁 Criando estrutura ADAPTERS..."

# HTTP Adapters
mkdir -p internal/adapters/http/handlers
touch internal/adapters/http/handlers/session_handler.go
touch internal/adapters/http/handlers/message_handler.go
touch internal/adapters/http/handlers/contact_handler.go
touch internal/adapters/http/handlers/group_handler.go
touch internal/adapters/http/handlers/chatwoot_handler.go
touch internal/adapters/http/handlers/webhook_handler.go
touch internal/adapters/http/handlers/media_handler.go
touch internal/adapters/http/handlers/newsletter_handler.go
touch internal/adapters/http/handlers/community_handler.go
touch internal/adapters/http/handlers/common_handler.go

mkdir -p internal/adapters/http/middleware
touch internal/adapters/http/middleware/auth.go
touch internal/adapters/http/middleware/cors.go
touch internal/adapters/http/middleware/logging.go
touch internal/adapters/http/middleware/validation.go
touch internal/adapters/http/middleware/rate_limit.go

mkdir -p internal/adapters/http/routes
touch internal/adapters/http/routes/routes.go
touch internal/adapters/http/routes/swagger.go

# Database Adapters
mkdir -p internal/adapters/database/postgres
touch internal/adapters/database/postgres/session_repository.go
touch internal/adapters/database/postgres/message_repository.go
touch internal/adapters/database/postgres/contact_repository.go
touch internal/adapters/database/postgres/group_repository.go
touch internal/adapters/database/postgres/chatwoot_repository.go
touch internal/adapters/database/postgres/webhook_repository.go
touch internal/adapters/database/postgres/media_repository.go
touch internal/adapters/database/postgres/newsletter_repository.go
touch internal/adapters/database/postgres/community_repository.go

mkdir -p internal/adapters/database/migrations
touch internal/adapters/database/migrations/001_initial_schema.up.sql
touch internal/adapters/database/migrations/001_initial_schema.down.sql
touch internal/adapters/database/migrations/002_add_indexes.up.sql
touch internal/adapters/database/migrations/002_add_indexes.down.sql

# WhatsApp Adapter
mkdir -p internal/adapters/whatsapp
touch internal/adapters/whatsapp/gateway.go
touch internal/adapters/whatsapp/client.go
touch internal/adapters/whatsapp/events.go
touch internal/adapters/whatsapp/mapper.go
touch internal/adapters/whatsapp/validator.go

# Chatwoot Adapter
mkdir -p internal/adapters/chatwoot
touch internal/adapters/chatwoot/gateway.go
touch internal/adapters/chatwoot/client.go
touch internal/adapters/chatwoot/webhook.go
touch internal/adapters/chatwoot/mapper.go
touch internal/adapters/chatwoot/formatter.go

# Events Adapter
mkdir -p internal/adapters/events
touch internal/adapters/events/publisher.go
touch internal/adapters/events/handlers.go
touch internal/adapters/events/dispatcher.go

# ==============================================
# 🚀 CMD - Entry Points
# ==============================================
echo "📁 Criando estrutura CMD..."

mkdir -p cmd/server
touch cmd/server/main.go
touch cmd/server/config.go
touch cmd/server/container.go

mkdir -p cmd/worker
touch cmd/worker/main.go
touch cmd/worker/jobs.go

mkdir -p cmd/cli
touch cmd/cli/main.go
touch cmd/cli/commands.go

# ==============================================
# 🏗️ PLATFORM - Infrastructure (complementar)
# ==============================================
echo "📁 Complementando estrutura PLATFORM..."

mkdir -p platform/container
touch platform/container/container.go
touch platform/container/wire.go

mkdir -p platform/monitoring
touch platform/monitoring/metrics.go
touch platform/monitoring/health.go

# ==============================================
# 🧪 TESTS - Test Structure
# ==============================================
echo "📁 Criando estrutura de TESTES..."

# Core Tests
mkdir -p tests/internal/core/session
touch tests/internal/core/session/service_test.go
touch tests/internal/core/session/models_test.go

mkdir -p tests/internal/core/messaging
touch tests/internal/core/messaging/service_test.go
touch tests/internal/core/messaging/models_test.go

# Services Tests
mkdir -p tests/internal/services
touch tests/internal/services/session_service_test.go
touch tests/internal/services/message_service_test.go

# Adapters Tests
mkdir -p tests/internal/adapters/http
touch tests/internal/adapters/http/handlers_test.go

mkdir -p tests/internal/adapters/database
touch tests/internal/adapters/database/repositories_test.go

# Integration Tests
mkdir -p tests/integration
touch tests/integration/api_test.go
touch tests/integration/database_test.go

# E2E Tests
mkdir -p tests/e2e
touch tests/e2e/session_flow_test.go
touch tests/e2e/message_flow_test.go

# Test Helpers
mkdir -p tests/helpers
touch tests/helpers/fixtures.go
touch tests/helpers/mocks.go
touch tests/helpers/testcontainers.go

# ==============================================
# 📋 SCRIPTS - Automation
# ==============================================
echo "📁 Criando estrutura de SCRIPTS..."

mkdir -p scripts
touch scripts/validate-architecture.sh
touch scripts/check-dependencies.sh
touch scripts/test-coverage.sh
touch scripts/generate-mocks.sh
touch scripts/run-migrations.sh

# ==============================================
# 📚 DOCS - Documentation
# ==============================================
echo "📁 Criando estrutura de DOCUMENTAÇÃO..."

mkdir -p docs/architecture
touch docs/architecture/core.md
touch docs/architecture/services.md
touch docs/architecture/adapters.md
touch docs/architecture/testing.md

mkdir -p docs/api
touch docs/api/session.md
touch docs/api/messaging.md
touch docs/api/contacts.md

mkdir -p docs/development
touch docs/development/setup.md
touch docs/development/testing.md
touch docs/development/deployment.md

# ==============================================
# ⚙️ CONFIG - Configuration Files
# ==============================================
echo "📁 Criando arquivos de CONFIGURAÇÃO..."

# Go Module files
touch go.work
touch .golangci.yml

# Test configuration
touch .testcoverage.yml

# Docker files
touch .dockerignore

# Git files
touch .gitignore

# Air configuration for hot reload
touch .air.toml

echo ""
echo "✅ Estrutura da Nova Arquitetura Clean Architecture criada com sucesso!"
echo ""
echo "📊 Resumo:"
echo "   🎯 Core: $(find internal/core -name "*.go" 2>/dev/null | wc -l) arquivos Go"
echo "   🔧 Services: $(find internal/services -name "*.go" 2>/dev/null | wc -l) arquivos Go"
echo "   🔌 Adapters: $(find internal/adapters -name "*.go" 2>/dev/null | wc -l) arquivos Go"
echo "   🚀 CMD: $(find cmd -name "*.go" 2>/dev/null | wc -l) arquivos Go"
echo "   🧪 Tests: $(find tests -name "*.go" 2>/dev/null | wc -l) arquivos de teste"
echo ""
echo "🔄 Próximos passos:"
echo "   1. Executar: go mod init zpwoot (se necessário)"
echo "   2. Começar migração do módulo Session"
echo "   3. Implementar testes unitários"
echo "   4. Configurar CI/CD"
echo ""
echo "📖 Consulte MIGRATION_PLAN.md para detalhes da migração"
