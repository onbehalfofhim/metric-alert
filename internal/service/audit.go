package service

import (
	"github.com/onbehalfofhim/metric-alert/internal/audit"
	"github.com/onbehalfofhim/metric-alert/internal/models"
)

// AuditPublisher - управляет наблюдателями
type AuditPublisher interface {
	Register(audit.AuditObserver)
	Deregister(audit.AuditObserver)
	Notify(models.AuditMessage)
}
