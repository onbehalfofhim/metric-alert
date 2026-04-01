package file

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"time"

	"github.com/onbehalfofhim/metric-alert/internal/logger"
	"github.com/onbehalfofhim/metric-alert/internal/models"
	"github.com/onbehalfofhim/metric-alert/internal/repository"
)

type FileStorage struct {
	storage repository.Storage
	file    *os.File
	logger  *logger.Logger
}

func NewFileStorage(storage repository.Storage, filePath string, logger *logger.Logger) (*FileStorage, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &FileStorage{
		storage: storage,
		file:    file,
		logger:  logger,
	}, nil
}

func (fs *FileStorage) save(metrics []models.Metric) error {
	// очищаем файл
	if err := fs.file.Truncate(0); err != nil {
		return err
	}

	if _, err := fs.file.Seek(0, 0); err != nil {
		return err
	}

	if err := json.NewEncoder(fs.file).Encode(metrics); err != nil {
		return err
	}

	return nil
}

func (fs *FileStorage) load() ([]models.Metric, error) {
	var metrics []models.Metric

	if _, err := fs.file.Seek(0, 0); err != nil {
		return nil, err
	}

	if err := json.NewDecoder(fs.file).Decode(&metrics); err != nil {
		if errors.Is(err, io.EOF) {
			return metrics, nil // пустой файл — это ок
		}
		return nil, err
	}

	return metrics, nil
}

func (fs *FileStorage) LoadFromFile() error {
	metrics, err := fs.load()
	if err != nil {
		return err
	}

	for _, m := range metrics {
		switch m.MType {
		case "gauge":
			err := fs.storage.UpdateGauge(m.ID, *m.Value)
			if err != nil {
				return err
			}
		case "counter":
			err := fs.storage.UpdateCounter(m.ID, *m.Delta)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (fs *FileStorage) GetMetrics() []models.Metric {
	gauges := fs.storage.GetListGauges()
	counters := fs.storage.GetListCounters()

	var metrics []models.Metric
	for k, v := range gauges {
		metrics = append(metrics, models.NewGauge(k, v))
	}
	for k, v := range counters {
		metrics = append(metrics, models.NewCounter(k, v))
	}

	return metrics
}

func (fs *FileStorage) SaveToFile() error {
	metrics := fs.GetMetrics()

	return fs.save(metrics)
}

func (fs *FileStorage) RunBackup(interval time.Duration) {
	ticker := time.NewTicker(interval)

	go func() {
		for range ticker.C {
			if err := fs.SaveToFile(); err != nil {
				fs.logger.Error("failed to save metrics", "error", err)
			}
		}
	}()
}

func (fs *FileStorage) Close() error {
	return fs.file.Close()
}

func (fs *FileStorage) Ping(ctx context.Context) error {
	return nil
}
