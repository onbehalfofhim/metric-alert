package audit

import "github.com/onbehalfofhim/metric-alert/internal/models"

// AuditObserver - реагирует на события аудита
type AuditObserver interface {
	GetID() string
	Notify(models.AuditMessage)
}
