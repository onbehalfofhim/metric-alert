package file

import (
	"encoding/json"
	"os"

	"github.com/onbehalfofhim/metric-alert/internal/models"
)

type FileStorage struct {
	path string
}

func NewFileStorage(filename string) *FileStorage {
	return &FileStorage{
		path: filename,
	}
}

func (fs *FileStorage) Save(metrics []models.Metric) error {
	file, err := os.Create(fs.path)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)

	if err := enc.Encode(metrics); err != nil {
		return err
	}

	return nil
}

func (f *FileStorage) Load() ([]models.Metric, error) {
	file, err := os.Open(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			// файла нет — это не ошибка, просто пустые данные
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var metrics []models.Metric

	dec := json.NewDecoder(file)
	if err := dec.Decode(&metrics); err != nil {
		return nil, err
	}

	return metrics, nil
}
