package storage

import (
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"os"
	"path"
	"time"
)

type fileProvider struct {
	filePath      string
	storeInterval int
	st            *MemStorage
}

func (f *fileProvider) Check() error {
	return errors.New("not provided for this storage type")
}

func NewFileProvider(filePath string, storeInterval int, m *MemStorage) StorageWorker {
	return &fileProvider{
		filePath:      filePath,
		storeInterval: storeInterval,
		st:            m,
	}
}

func (f *fileProvider) Dump() error {
	dir, _ := path.Split(f.filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0666)
		if err != nil {
			zap.S().Error(err)
		}
	}

	data, err := json.MarshalIndent(f.st, "", "   ")
	if err != nil {
		return err
	}

	return os.WriteFile(f.filePath, data, 0666)
}

func (f *fileProvider) IntervalDump() {
	pollTicker := time.NewTicker(time.Duration(f.storeInterval) * time.Second)
	defer pollTicker.Stop()
	for range pollTicker.C {
		err := f.Dump()
		if err != nil {
			zap.S().Error(err)
		}
	}
}

func (f *fileProvider) Restore() error {
	file, err := os.ReadFile(f.filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(file, f.st)
}
