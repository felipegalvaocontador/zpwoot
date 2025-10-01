# Chatwoot Integration

Esta implementação fornece integração bidirecional entre WhatsApp (via wameow) e Chatwoot, baseada na Evolution API.

## 🏗️ Arquitetura

### Componentes Principais

1. **Client** (`client.go`) - Cliente HTTP para API Chatwoot
2. **Manager** (`manager.go`) - Gerenciador de clientes por sessão
3. **WebhookHandler** (`webhook.go`) - Processamento de webhooks Chatwoot
4. **MessageMapper** (`message_mapper.go`) - Mapeamento de IDs de mensagens
5. **IntegrationManager** (`integration_manager.go`) - Coordenação da integração
6. **ContactSync** (`contact.go`) - Sincronização de contatos
7. **ConversationManager** (`conversation.go`) - Gestão de conversas
8. **MessageFormatter** (`formatter.go`) - Formatação markdown bidirecional
9. **ImportManager** (`import.go`) - Importação de histórico
10. **Utils** (`utils.go`) - Utilitários e validações

### Fluxo de Dados

```
WhatsApp Message → wameow/events.go → IntegrationManager → Chatwoot API
                                   ↓
                              MessageMapper (zpMessage table)
                                   ↓
                              Mapping: zpMessageId ↔ cwMessageId
```

## 🗃️ Estrutura de Banco

### Tabela zpMessage (Mapeamento Simplificado)

```sql
CREATE TABLE "zpMessage" (
    id VARCHAR PRIMARY KEY,
    "sessionId" VARCHAR NOT NULL,
    "zpMessageId" VARCHAR NOT NULL,      -- WhatsApp message ID
    "cwMessageId" INTEGER,               -- Chatwoot message ID
    "cwConversationId" INTEGER,          -- Chatwoot conversation ID
    "syncStatus" VARCHAR NOT NULL,       -- pending, synced, failed
    "createdAt" TIMESTAMP NOT NULL,
    "updatedAt" TIMESTAMP NOT NULL,
    "syncedAt" TIMESTAMP
);
```

**Nota:** Esta tabela serve apenas para mapeamento de IDs. Não armazenamos o conteúdo das mensagens do Chatwoot.

## 🔧 Configuração

### 1. Configuração Chatwoot

```go
config := &chatwoot.ChatwootConfig{
    URL:       "http://localhost:3001",
    APIKey:    "WAF6y4K5s6sdR9uVpsdE7BCt",
    AccountID: "1",
    Active:    true,
}
```

### 2. Inicialização

```go
// Criar componentes
logger := logger.New("chatwoot")
client := chatwoot.NewClient(config.URL, config.APIKey, config.AccountID, logger)
manager := chatwoot.NewManager(logger, repository)
messageMapper := chatwoot.NewMessageMapper(logger, zpMessageRepo)

// Criar integration manager
integrationMgr := chatwoot.NewIntegrationManager(
    logger, manager, messageMapper, contactSync, conversationMgr, formatter,
)

// Conectar ao wameow
eventHandler.SetChatwootManager(integrationMgr)
```

## 📨 Fluxo de Mensagens

### WhatsApp → Chatwoot

1. **Evento recebido** no `wameow/events.go`
2. **Verificação** se Chatwoot está habilitado
3. **Criação de mapeamento** na tabela `zpMessage`
4. **Extração** de dados da mensagem (phone, content, type)
5. **Criação/busca** de contato no Chatwoot
6. **Criação/busca** de conversa no Chatwoot
7. **Formatação** do conteúdo (markdown)
8. **Envio** da mensagem para Chatwoot
9. **Atualização** do mapeamento com IDs do Chatwoot

### Chatwoot → WhatsApp

1. **Webhook recebido** do Chatwoot
2. **Processamento** no `WebhookHandler`
3. **Filtros** aplicados (mensagens privadas, bot, etc.)
4. **Busca** do mapeamento na tabela `zpMessage`
5. **Formatação** do conteúdo para WhatsApp
6. **Envio** via wameow para WhatsApp

## 🎯 Funcionalidades Implementadas

### ✅ Básicas
- [x] Cliente HTTP Chatwoot completo
- [x] Mapeamento de mensagens (zpMessage)
- [x] Sincronização WhatsApp → Chatwoot
- [x] Formatação markdown bidirecional
- [x] Gestão de contatos e conversas
- [x] Sistema de webhook

### ✅ Avançadas
- [x] Normalização de números brasileiros
- [x] Filtros de mensagens (privadas, bot)
- [x] Retry com backoff exponencial
- [x] Logs estruturados
- [x] Validações de URL/token
- [x] Estatísticas de mapeamento

### ⚠️ Em Desenvolvimento
- [ ] Processamento completo de webhooks Chatwoot → WhatsApp
- [ ] Importação de histórico
- [ ] Funcionalidades específicas Evolution API
- [ ] Testes unitários e integração

## 🔍 Monitoramento

### Logs Estruturados

```go
logger.InfoWithFields("Message processed", map[string]interface{}{
    "session_id":         sessionID,
    "message_id":         messageID,
    "cw_message_id":      chatwootMessage.ID,
    "cw_conversation_id": conversation.ID,
})
```

### Estatísticas

```go
stats, err := integrationMgr.GetMappingStats(sessionID)
// Returns: {Total: 150, Pending: 5, Synced: 140, Failed: 5}
```

## 🧪 Testes

### Teste Manual

```bash
# Verificar Chatwoot rodando
curl -H "api_access_token: WAF6y4K5s6sdR9uVpsdE7BCt" \
     http://localhost:3001/api/v1/accounts/1

# Enviar mensagem WhatsApp
# Verificar se aparece no Chatwoot

# Responder no Chatwoot
# Verificar se chega no WhatsApp
```

### Verificar Mapeamentos

```sql
SELECT * FROM "zpMessage" WHERE "sessionId" = 'your-session-id';
```

## 🚨 Troubleshooting

### Problemas Comuns

1. **Mensagens não aparecem no Chatwoot**
   - Verificar se Chatwoot está habilitado: `IsEnabled(sessionID)`
   - Verificar logs de erro na criação de contato/conversa
   - Verificar token e URL do Chatwoot

2. **Mapeamentos ficam "pending"**
   - Verificar conectividade com Chatwoot
   - Verificar se inbox existe
   - Executar `ProcessPendingMessages()`

3. **Webhooks não funcionam**
   - Verificar URL do webhook configurada no Chatwoot
   - Verificar se endpoint está acessível
   - Verificar logs do WebhookHandler

### Debug

```go
// Habilitar logs debug
logger.SetLevel("debug")

// Verificar stats
stats, _ := integrationMgr.GetMappingStats(sessionID)
fmt.Printf("Stats: %+v\n", stats)

// Processar mensagens pendentes
err := integrationMgr.ProcessPendingMessages(sessionID, 10)
```

## 📚 Referências

- [Evolution API](https://github.com/EvolutionAPI/evolution-api) - Referência de implementação
- [Chatwoot API](https://www.chatwoot.com/developers/api/) - Documentação oficial
- [WhatsApp Business API](https://developers.facebook.com/docs/whatsapp) - Especificações

## 🔄 Próximos Passos

1. **Completar webhook handler** Chatwoot → WhatsApp
2. **Implementar importação** de histórico
3. **Adicionar testes** unitários e integração
4. **Otimizar performance** com cache
5. **Adicionar métricas** detalhadas
