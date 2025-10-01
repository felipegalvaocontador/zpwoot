package services

import (
	"zpwoot/platform/logger"
)

// ContactService implementa operações de contatos integradas ao WhatsApp
type ContactService struct {
	logger *logger.Logger
}

// NewContactService cria nova instância do serviço de contatos
func NewContactService(logger *logger.Logger) *ContactService {
	return &ContactService{
		logger: logger,
	}
}