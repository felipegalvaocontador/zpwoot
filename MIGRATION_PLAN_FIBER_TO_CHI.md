# Plano de Migração: Fiber para Chi - zpwoot

## 📋 Resumo Executivo

Este documento detalha o plano completo para migrar a aplicação zpwoot do framework **Fiber** para **Chi**, mantendo 100% da funcionalidade existente, todos os endpoints, middlewares e padrões de resposta.

## 🎯 Objetivos

- ✅ Migrar completamente do Fiber para Chi
- ✅ Manter todos os endpoints existentes
- ✅ Preservar todos os middlewares
- ✅ Manter compatibilidade de API
- ✅ Não alterar estrutura de rotas ou endpoints
- ✅ Preservar autenticação e logging

## 📊 Análise da Estrutura Atual

### Endpoints Mapeados

#### Health Endpoints
- `GET /health` - Health check básico
- `GET /health/wameow` - Health check do WhatsApp manager

#### Session Management
- `POST /sessions/create` - Criar nova sessão
- `GET /sessions/list` - Listar sessões
- `GET /sessions/:sessionId/info` - Info da sessão
- `DELETE /sessions/:sessionId/delete` - Deletar sessão
- `POST /sessions/:sessionId/connect` - Conectar sessão
- `POST /sessions/:sessionId/logout` - Logout da sessão
- `GET /sessions/:sessionId/qr` - Obter QR Code
- `POST /sessions/:sessionId/pair` - Pareamento por telefone
- `POST /sessions/:sessionId/proxy/set` - Configurar proxy
- `GET /sessions/:sessionId/proxy/find` - Obter proxy

#### Message Operations
- `POST /sessions/:sessionId/messages/send/text` - Enviar texto
- `POST /sessions/:sessionId/messages/send/media` - Enviar mídia
- `POST /sessions/:sessionId/messages/send/image` - Enviar imagem
- `POST /sessions/:sessionId/messages/send/audio` - Enviar áudio
- `POST /sessions/:sessionId/messages/send/video` - Enviar vídeo
- `POST /sessions/:sessionId/messages/send/document` - Enviar documento
- `POST /sessions/:sessionId/messages/send/sticker` - Enviar sticker
- `POST /sessions/:sessionId/messages/send/button` - Enviar botões
- `POST /sessions/:sessionId/messages/send/contact` - Enviar contato
- `POST /sessions/:sessionId/messages/send/list` - Enviar lista
- `POST /sessions/:sessionId/messages/send/location` - Enviar localização
- `POST /sessions/:sessionId/messages/send/poll` - Enviar enquete
- `POST /sessions/:sessionId/messages/send/reaction` - Enviar reação
- `POST /sessions/:sessionId/messages/send/presence` - Enviar presença
- `POST /sessions/:sessionId/messages/edit` - Editar mensagem
- `POST /sessions/:sessionId/messages/mark-read` - Marcar como lida
- `POST /sessions/:sessionId/messages/revoke` - Revogar mensagem
- `GET /sessions/:sessionId/messages/poll/:messageId/results` - Resultados da enquete

#### Group Management
- `POST /sessions/:sessionId/groups/create` - Criar grupo
- `GET /sessions/:sessionId/groups` - Listar grupos
- `GET /sessions/:sessionId/groups/info` - Info do grupo
- `POST /sessions/:sessionId/groups/participants` - Gerenciar participantes
- `PUT /sessions/:sessionId/groups/name` - Alterar nome
- `PUT /sessions/:sessionId/groups/description` - Alterar descrição
- `PUT /sessions/:sessionId/groups/photo` - Alterar foto
- `GET /sessions/:sessionId/groups/invite-link` - Link de convite
- `POST /sessions/:sessionId/groups/join` - Entrar no grupo
- `POST /sessions/:sessionId/groups/leave` - Sair do grupo
- `PUT /sessions/:sessionId/groups/settings` - Configurações
- `GET /sessions/:sessionId/groups/requests` - Solicitações pendentes
- `POST /sessions/:sessionId/groups/requests` - Gerenciar solicitações
- `PUT /sessions/:sessionId/groups/join-approval` - Modo de aprovação
- `PUT /sessions/:sessionId/groups/member-add-mode` - Modo de adição
- `GET /sessions/:sessionId/groups/info-from-link` - Info por link
- `POST /sessions/:sessionId/groups/info-from-invite` - Info por convite
- `POST /sessions/:sessionId/groups/join-with-invite` - Entrar por convite

#### Newsletter Management
- `POST /sessions/:sessionId/newsletters/create` - Criar newsletter
- `GET /sessions/:sessionId/newsletters/info` - Info da newsletter
- `POST /sessions/:sessionId/newsletters/info-from-invite` - Info por convite
- `POST /sessions/:sessionId/newsletters/follow` - Seguir newsletter
- `POST /sessions/:sessionId/newsletters/unfollow` - Parar de seguir
- `GET /sessions/:sessionId/newsletters/messages` - Mensagens
- `GET /sessions/:sessionId/newsletters/updates` - Atualizações
- `POST /sessions/:sessionId/newsletters/mark-viewed` - Marcar como vista
- `POST /sessions/:sessionId/newsletters/send-reaction` - Enviar reação
- `POST /sessions/:sessionId/newsletters/subscribe-live` - Inscrever em atualizações
- `POST /sessions/:sessionId/newsletters/toggle-mute` - Alternar silenciar
- `POST /sessions/:sessionId/newsletters/accept-tos` - Aceitar termos
- `POST /sessions/:sessionId/newsletters/upload` - Upload newsletter
- `POST /sessions/:sessionId/newsletters/upload-reader` - Upload reader
- `GET /sessions/:sessionId/newsletters` - Newsletters inscritas

#### Community Management
- `POST /sessions/:sessionId/communities/link-group` - Vincular grupo
- `POST /sessions/:sessionId/communities/unlink-group` - Desvincular grupo
- `GET /sessions/:sessionId/communities/info` - Info da comunidade
- `GET /sessions/:sessionId/communities/subgroups` - Sub-grupos

#### Contact Management
- `POST /sessions/:sessionId/contacts/check` - Verificar WhatsApp
- `GET /sessions/:sessionId/contacts/avatar` - Foto do perfil
- `POST /sessions/:sessionId/contacts/info` - Info do usuário
- `GET /sessions/:sessionId/contacts` - Listar contatos
- `POST /sessions/:sessionId/contacts/sync` - Sincronizar contatos
- `GET /sessions/:sessionId/contacts/business` - Perfil comercial

#### Webhook Management
- `POST /sessions/:sessionId/webhook/set` - Configurar webhook
- `GET /sessions/:sessionId/webhook/find` - Obter webhook
- `POST /sessions/:sessionId/webhook/test` - Testar webhook
- `GET /webhook/events` - Eventos suportados

#### Chatwoot Integration
- `POST /sessions/:sessionId/chatwoot/set` - Configurar Chatwoot
- `GET /sessions/:sessionId/chatwoot/find` - Obter configuração
- `POST /sessions/:sessionId/chatwoot/contacts/sync` - Sincronizar contatos
- `POST /sessions/:sessionId/chatwoot/conversations/sync` - Sincronizar conversas
- `POST /sessions/:sessionId/chatwoot/webhook` - Webhook Chatwoot
- `POST /chatwoot/webhook/:sessionId` - Webhook alternativo

#### Swagger Documentation
- `GET /swagger/*` - Documentação da API

### Middlewares Identificados

1. **RequestID** - Geração de ID único por requisição
2. **HTTPLogger** - Log detalhado de requisições HTTP
3. **Metrics** - Coleta de métricas de requisições
4. **APIKeyAuth** - Autenticação por API Key
5. **Recover** - Recovery de panics
6. **CORS** - Cross-Origin Resource Sharing

### Dependências do Fiber

#### Dependências Diretas (go.mod)
```go
github.com/gofiber/fiber/v2 v2.52.9
github.com/swaggo/fiber-swagger v1.3.0
```

#### Dependências Indiretas (relacionadas ao Fiber)
```go
github.com/andybalholm/brotli v1.2.0 // indirect
github.com/klauspost/compress v1.18.0 // indirect
github.com/mattn/go-colorable v0.1.14 // indirect
github.com/mattn/go-isatty v0.0.20 // indirect
github.com/mattn/go-runewidth v0.0.17 // indirect
github.com/rivo/uniseg v0.4.7 // indirect
github.com/swaggo/files v1.0.1 // indirect
github.com/valyala/bytebufferpool v1.0.0 // indirect
github.com/valyala/fasthttp v1.66.0 // indirect
```

#### Arquivos com Imports do Fiber
- `cmd/zpwoot/main.go` - Servidor principal e middlewares
- `platform/logger/middleware.go` - Logger middleware
- `internal/infra/http/routers/routes.go` - Sistema de rotas
- `internal/infra/http/middleware/*.go` - Todos os middlewares customizados
- `internal/infra/http/handlers/*.go` - Todos os handlers HTTP

#### Middlewares Built-in do Fiber Utilizados
- `github.com/gofiber/fiber/v2/middleware/cors` - CORS
- `github.com/gofiber/fiber/v2/middleware/recover` - Recovery de panics
- `github.com/gofiber/fiber/v2/middleware/logger` - Logger (usado em platform/logger/middleware.go)

#### Swagger Integration
- `github.com/swaggo/fiber-swagger` - Integração Swagger para Fiber

## 🔄 Mapeamento Fiber -> Chi

### Padrões de Uso Identificados no Codebase

#### Métodos do Contexto Fiber Mais Utilizados
1. **`c.Params("sessionId")`** - Extração de parâmetros de rota (usado em todos os handlers)
2. **`c.BodyParser(&struct)`** - Parse do body JSON (usado extensivamente)
3. **`c.JSON(response)`** - Retorno de respostas JSON (padrão em todos os endpoints)
4. **`c.Status(code).JSON(data)`** - Definição de status HTTP com resposta JSON
5. **`c.Query("param")`** - Extração de query parameters
6. **`c.Context()`** - Obtenção do context.Context para use cases
7. **`c.Get("header")`** - Leitura de headers (Authorization, User-Agent, etc.)
8. **`c.Set("header", "value")`** - Definição de headers de resposta
9. **`c.Locals("key")`** - Armazenamento/recuperação de valores locais
10. **`c.IP()`**, **`c.Method()`**, **`c.Path()`** - Informações da requisição para logs

#### Padrões de Resposta Identificados
- **Sucesso**: `c.JSON(response)` ou `c.Status(200).JSON(response)`
- **Criação**: `c.Status(201).JSON(response)`
- **Erro 400**: `c.Status(400).JSON(common.NewErrorResponse("message"))`
- **Erro 404**: `c.Status(404).JSON(common.NewErrorResponse("message"))`
- **Erro 500**: `c.Status(500).JSON(common.NewErrorResponse("message"))`
- **Fiber.Map**: `c.Status(code).JSON(fiber.Map{"key": "value"})`

### Conceitos Principais

| Fiber | Chi | Descrição |
|-------|-----|-----------|
| `fiber.App` | `chi.Router` | Router principal |
| `fiber.Ctx` | `http.ResponseWriter + *http.Request` | Contexto da requisição |
| `c.Params("key")` | `chi.URLParam(r, "key")` | Parâmetros de rota |
| `c.Query("key")` | `r.URL.Query().Get("key")` | Query parameters |
| `c.QueryBool("key", default)` | `r.URL.Query().Get("key") == "true"` | Query boolean |
| `c.BodyParser(&struct)` | `json.NewDecoder(r.Body).Decode(&struct)` | Parse do body |
| `c.Body()` | `io.ReadAll(r.Body)` | Body raw |
| `c.JSON(data)` | `json.NewEncoder(w).Encode(data)` | Resposta JSON |
| `c.Status(code)` | `w.WriteHeader(code)` | Status HTTP |
| `c.Status(code).JSON(data)` | `w.WriteHeader(code) + json.NewEncoder(w).Encode(data)` | Status + JSON |
| `c.Get("header")` | `r.Header.Get("header")` | Headers da requisição |
| `c.Set("header", "value")` | `w.Header().Set("header", "value")` | Headers da resposta |
| `c.IP()` | `r.RemoteAddr` | IP do cliente |
| `c.Method()` | `r.Method` | Método HTTP |
| `c.Path()` | `r.URL.Path` | Caminho da URL |
| `c.Context()` | `r.Context()` | Context do Go |
| `c.Locals("key")` | `r.Context().Value("key")` | Valores locais |
| `c.Locals("key", value)` | `context.WithValue(r.Context(), "key", value)` | Definir valores locais |
| `c.Next()` | `next.ServeHTTP(w, r)` | Próximo middleware |
| `fiber.Map{}` | `map[string]interface{}{}` | Map genérico |

### Middlewares

| Fiber | Chi | Descrição |
|-------|-----|-----------|
| `recover.New()` | `middleware.Recoverer` | Recovery de panics |
| `cors.New()` | `cors.Handler()` | CORS |
| `app.Use(middleware)` | `r.Use(middleware)` | Aplicar middleware |
| `app.Group("/path")` | `r.Route("/path", func(r chi.Router) {...})` | Agrupamento de rotas |

### Dependências Equivalentes

| Fiber | Chi | Descrição |
|-------|-----|-----------|
| `github.com/gofiber/fiber/v2` | `github.com/go-chi/chi/v5` | Framework principal |
| `github.com/swaggo/fiber-swagger` | `github.com/swaggo/http-swagger` | Swagger integration |
| `github.com/gofiber/fiber/v2/middleware/cors` | `github.com/go-chi/cors` | CORS middleware |
| `github.com/gofiber/fiber/v2/middleware/recover` | `github.com/go-chi/chi/v5/middleware` | Recovery middleware |
| `github.com/gofiber/fiber/v2/middleware/logger` | `github.com/go-chi/chi/v5/middleware` | Logger middleware |

### Estratégias de Migração

#### 1. Handlers
- Converter `func(c *fiber.Ctx) error` para `func(w http.ResponseWriter, r *http.Request)`
- Criar helpers para extrair parâmetros, query strings e body
- Implementar helpers para respostas JSON padronizadas
- Manter estruturas de resposta existentes (common.SuccessResponse, common.ErrorResponse)

#### 2. Middlewares
- Converter middlewares customizados para assinatura Chi
- Usar middlewares built-in do Chi quando possível
- Manter funcionalidade de logging e métricas

#### 3. Rotas
- Converter `app.Get/Post/Put/Delete` para `r.Get/Post/Put/Delete`
- Manter estrutura de agrupamento com `r.Route`
- Preservar todos os parâmetros de rota (`:sessionId`, etc.)

#### 4. Error Handling
- Implementar middleware de error handling global
- Manter padrões de resposta de erro existentes
- Preservar códigos de status HTTP

## 📁 Arquivos a Modificar

### 1. Arquivo Principal
- `cmd/zpwoot/main.go` - Configuração do servidor

### 2. Middlewares
- `internal/infra/http/middleware/request_id.go`
- `internal/infra/http/middleware/logger.go`
- `internal/infra/http/middleware/metrics.go`
- `internal/infra/http/middleware/auth.go`

### 3. Handlers
- `internal/infra/http/handlers/health.go`
- `internal/infra/http/handlers/session.go`
- `internal/infra/http/handlers/message.go`
- `internal/infra/http/handlers/group.go`
- `internal/infra/http/handlers/contact.go`
- `internal/infra/http/handlers/newsletter.go`
- `internal/infra/http/handlers/community.go`
- `internal/infra/http/handlers/webhook.go`
- `internal/infra/http/handlers/chatwoot.go`
- `internal/infra/http/handlers/media.go`
- `internal/infra/http/handlers/common.go`

### 4. Rotas
- `internal/infra/http/routers/routes.go`

### 5. Helpers
- `internal/infra/http/helpers/session_resolver.go`

### 6. Configuração
- `go.mod` - Dependências
- `.air.toml` - Hot reload (verificar compatibilidade)

### Helpers Necessários

#### Chi Context Helpers
```go
// Extrair parâmetros de rota
func GetURLParam(r *http.Request, key string) string {
    return chi.URLParam(r, key)
}

// Extrair query parameters
func GetQueryParam(r *http.Request, key string) string {
    return r.URL.Query().Get(key)
}

// Parse JSON body
func ParseJSONBody(r *http.Request, v interface{}) error {
    return json.NewDecoder(r.Body).Decode(v)
}

// Resposta JSON
func WriteJSON(w http.ResponseWriter, status int, data interface{}) error {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    return json.NewEncoder(w).Encode(data)
}
```

## 🚀 Próximos Passos

1. **Análise Completa** ✅ (Concluído)
2. **Atualização de Dependências**
3. **Migração do Servidor Principal**
4. **Migração dos Middlewares**
5. **Migração dos Handlers**
6. **Migração do Sistema de Rotas**
7. **Migração dos Helpers**
8. **Testes e Validação**
9. **Documentação e Limpeza**

## 📈 Estimativa de Esforço

- **Total de arquivos a modificar**: ~25 arquivos
- **Endpoints a migrar**: ~80 endpoints
- **Middlewares a migrar**: 5 middlewares customizados + 3 built-in
- **Tempo estimado**: 2-3 dias de desenvolvimento + 1 dia de testes

---

**Status**: 🔄 Em Progresso - Análise Completa
**Próxima Etapa**: Atualização de dependências no go.mod
