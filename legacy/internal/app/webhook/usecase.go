package webhook

import (
	"context"

	"zpwoot/internal/domain/webhook"
	"zpwoot/internal/ports"
)

type UseCase interface {
	SetConfig(ctx context.Context, req *SetConfigRequest) (*SetConfigResponse, error)
	FindConfig(ctx context.Context, sessionID string) (*WebhookResponse, error)
	UpdateWebhook(ctx context.Context, webhookID string, req *UpdateWebhookRequest) (*WebhookResponse, error)
	DeleteWebhook(ctx context.Context, webhookID string) error
	ListWebhooks(ctx context.Context, req *ListWebhooksRequest) (*ListWebhooksResponse, error)
	TestWebhook(ctx context.Context, webhookID string, req *TestWebhookRequest) (*TestWebhookResponse, error)
	GetSupportedWebhookEvents(ctx context.Context) (*WebhookEventsResponse, error)
	ProcessWebhookEvent(ctx context.Context, event *webhook.WebhookEvent) error
}

type useCaseImpl struct {
	webhookRepo    ports.WebhookRepository
	webhookService *webhook.Service
}

func NewUseCase(
	webhookRepo ports.WebhookRepository,
	webhookService *webhook.Service,
) UseCase {
	return &useCaseImpl{
		webhookRepo:    webhookRepo,
		webhookService: webhookService,
	}
}

func (uc *useCaseImpl) SetConfig(ctx context.Context, req *SetConfigRequest) (*SetConfigResponse, error) {
	domainReq := req.ToSetConfigRequest()

	webhookConfig, err := uc.webhookService.SetConfig(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	response := &SetConfigResponse{
		ID:        webhookConfig.ID.String(),
		SessionID: webhookConfig.SessionID,
		URL:       webhookConfig.URL,
		Events:    webhookConfig.Events,
		Enabled:   webhookConfig.Enabled,
		CreatedAt: webhookConfig.CreatedAt,
	}

	return response, nil
}

func (uc *useCaseImpl) FindConfig(ctx context.Context, sessionID string) (*WebhookResponse, error) {
	webhookConfig, err := uc.webhookService.GetWebhookBySession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	response := FromWebhook(webhookConfig)
	return response, nil
}

func (uc *useCaseImpl) UpdateWebhook(ctx context.Context, webhookID string, req *UpdateWebhookRequest) (*WebhookResponse, error) {
	domainReq := req.ToUpdateWebhookRequest()

	webhookConfig, err := uc.webhookService.UpdateWebhook(ctx, webhookID, domainReq)
	if err != nil {
		return nil, err
	}

	response := FromWebhook(webhookConfig)
	return response, nil
}

func (uc *useCaseImpl) DeleteWebhook(ctx context.Context, webhookID string) error {
	return uc.webhookService.DeleteWebhook(ctx, webhookID)
}

func (uc *useCaseImpl) ListWebhooks(ctx context.Context, req *ListWebhooksRequest) (*ListWebhooksResponse, error) {
	domainReq := req.ToListWebhooksRequest()

	if domainReq.Limit == 0 {
		domainReq.Limit = 20
	}

	webhooks, total, err := uc.webhookService.ListWebhooks(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	webhookResponses := make([]WebhookResponse, len(webhooks))
	for i, wh := range webhooks {
		webhookResponses[i] = *FromWebhook(wh)
	}

	response := &ListWebhooksResponse{
		Webhooks: webhookResponses,
		Total:    total,
		Limit:    domainReq.Limit,
		Offset:   domainReq.Offset,
	}

	return response, nil
}

func (uc *useCaseImpl) TestWebhook(ctx context.Context, webhookID string, req *TestWebhookRequest) (*TestWebhookResponse, error) {
	testEvent := &webhook.WebhookEvent{
		ID:        "test-" + webhookID,
		SessionID: "test-session",
		Type:      req.EventType,
		Data:      req.TestData,
	}

	result, err := uc.webhookService.TestWebhook(ctx, webhookID, testEvent)
	if err != nil {
		return &TestWebhookResponse{
			Success: false,
			Error:   err.Error(),
		}, err
	}

	response := &TestWebhookResponse{
		Success:      result.Success,
		StatusCode:   result.StatusCode,
		ResponseTime: result.ResponseTime,
	}

	if result.Error != nil {
		response.Error = result.Error.Error()
	}

	return response, nil
}

func (uc *useCaseImpl) GetSupportedWebhookEvents(ctx context.Context) (*WebhookEventsResponse, error) {
	return GetSupportedEvents(), nil
}

func (uc *useCaseImpl) ProcessWebhookEvent(ctx context.Context, event *webhook.WebhookEvent) error {
	return uc.webhookService.ProcessEvent(ctx, event)
}
