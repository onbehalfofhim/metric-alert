package audit

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/onbehalfofhim/metric-alert/internal/logger"
	"github.com/onbehalfofhim/metric-alert/internal/models"
)

// FileObserver - запись событий аудита в файл.
type FileObserver struct {
	path   string
	logger *logger.Logger
}

// NewFileObserver - создает экземпляр наблюдателя для аудита.
func NewFileObserver(filePath string, logger *logger.Logger) *FileObserver {
	return &FileObserver{
		path:   filePath,
		logger: logger,
	}
}

// GetID - получение идентификатора наблюдателя.
func (o *FileObserver) GetID() string {
	return fmt.Sprintf("file-observer-%s", o.path)
}

// Notify - реагирование наблюдателя на событие аудита.
func (o *FileObserver) Notify(message models.AuditMessage) {
	if err := o.writeToFile(message); err != nil {
		o.logger.Info("failed to write audit message to file", err)
		return
	}
	o.logger.Info("audit message written to file", "path", o.path)
}

func (o *FileObserver) writeToFile(message models.AuditMessage) error {
	file, err := os.OpenFile(o.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			o.logger.Info("close file: %v", err)
		}
	}()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(message); err != nil {
		return err
	}

	return nil
}
