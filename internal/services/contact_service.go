package services

import (
	"zpwoot/platform/logger"
)

type ContactService struct {
	logger *logger.Logger
}

func NewContactService(logger *logger.Logger) *ContactService {
	return &ContactService{
		logger: logger,
	}
}
