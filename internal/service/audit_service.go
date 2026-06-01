package service

import (
	"github.com/onbehalfofhim/metric-alert/internal/audit"
	"github.com/onbehalfofhim/metric-alert/internal/logger"
	"github.com/onbehalfofhim/metric-alert/internal/models"
)

// AuditService - управляет наблюдателями
type AuditService struct {
	observers map[string]audit.AuditObserver
	logger    *logger.Logger
}

func NewAuditService(logger *logger.Logger) *AuditService {
	return &AuditService{
		observers: map[string]audit.AuditObserver{},
		logger:    logger,
	}
}

func (s *AuditService) Notify(message models.AuditMessage) {
	s.logger.Info("notifying all observers", "observers_count", len(s.observers))

	for _, o := range s.observers {
		s.logger.Info("notifying observer", "observer_id", o.GetID())
		o.Notify(message)
	}
}

func (s *AuditService) Register(o audit.AuditObserver) {
	if s.observers == nil {
		s.observers = make(map[string]audit.AuditObserver)
	}

	s.observers[o.GetID()] = o

	s.logger.Info("attaching observer", "observer_id", o.GetID())
}

func (s *AuditService) Deregister(o audit.AuditObserver) {
	delete(s.observers, o.GetID())
	s.logger.Info("detaching observer", "observer_id", o.GetID())
}
