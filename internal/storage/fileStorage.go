package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/DeneesK/metrics-alerting/internal/models"
	"github.com/sethvargo/go-retry"
	"go.uber.org/zap"
)

type allMetrics struct {
	Gauge   map[string]float64 `json:"gauge"`
	Counter map[string]int64   `json:"counter"`
}

type producer struct {
	file    *os.File
	encoder *json.Encoder
}

type FileStorage struct {
	memoryStorage *MemStorage
	filePath      string
	storeInterval time.Duration
	log           *zap.SugaredLogger
}

func NewFileStorage(filePath string, storeInterval int, isRestore bool, log *zap.SugaredLogger) (*FileStorage, error) {
	ms, err := NewMemStorage(log)
	if err != nil {
		return nil, fmt.Errorf("impossible to create new storage - %w", err)
	}
	fs := FileStorage{
		memoryStorage: ms,
		filePath:      filePath,
		storeInterval: time.Duration(storeInterval) * time.Second,
		log:           ms.log,
	}

	if isRestore {
		if err := fs.loadFromFile(filePath); err != nil {
			ctx := context.Background()
			b := retry.WithMaxRetries(3, NewlinearBackoff(time.Second*1))
			err = retry.Do(ctx, b, tryLoadFile(filePath, &fs))
			if err != nil {
				return nil, fmt.Errorf("unable to load metrics from file - %w", err)
			}
		}
	}
	if storeInterval != 0 {
		if err := fs.startStoring(); err != nil {
			ms.log.Debugf("during initializing of new storage, error occurred: %w", err)
			return nil, err
		}
	}

	return &fs, nil
}

func (storage *FileStorage) Store(typeMetric string, name string, value interface{}) error {
	return storage.memoryStorage.Store(typeMetric, name, value)
}

func (storage *FileStorage) StoreBatch(metrics []models.Metrics) error {
	return storage.memoryStorage.StoreBatch(metrics)
}

func (storage *FileStorage) GetCounterMetrics() (map[string]int64, error) {
	return storage.memoryStorage.GetCounterMetrics()
}

func (storage *FileStorage) GetGaugeMetrics() (map[string]float64, error) {
	return storage.memoryStorage.GetGaugeMetrics()
}

func (storage *FileStorage) GetValue(typeMetric, name string) (Result, bool, error) {
	return storage.memoryStorage.GetValue(typeMetric, name)
}

func (storage *FileStorage) Ping() error {
	return storage.memoryStorage.Ping()
}

func (storage *FileStorage) Close() error {
	return nil
}

func newProducer(filename string) (*producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &producer{file: file, encoder: json.NewEncoder(file)}, nil
}

func (p *producer) writeMetrics(m *allMetrics) error {
	return p.encoder.Encode(&m)
}

func (p *producer) close() error {
	return p.file.Close()
}

type consumer struct {
	file    *os.File
	decoder *json.Decoder
}

func newConsumer(filename string) (*consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	return &consumer{file: file, decoder: json.NewDecoder(file)}, nil
}

func (c *consumer) readMetrics() (*allMetrics, error) {
	data := allMetrics{make(map[string]float64, 0), make(map[string]int64, 0)}
	err := c.decoder.Decode(&data)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return &data, nil
}

func (c *consumer) close() error {
	return c.file.Close()
}

func (storage *FileStorage) saveToFile(path string) error {
	p, err := newProducer(path)
	if err != nil {
		return err
	}
	defer p.close()
	var metrics allMetrics

	metrics.Counter, err = storage.memoryStorage.GetCounterMetrics()
	if err != nil {
		return err
	}
	metrics.Gauge, err = storage.memoryStorage.GetGaugeMetrics()
	if err != nil {
		return err
	}
	return p.writeMetrics(&metrics)
}

func (storage *FileStorage) loadFromFile(path string) error {
	c, err := newConsumer(path)
	if err != nil {
		return err
	}
	defer c.close()
	metrics, err := c.readMetrics()
	if err != nil {
		return err
	}
	storage.memoryStorage.setMetrics(metrics)
	return nil
}

func (storage *FileStorage) startStoring() error {
	dir, _ := path.Split(storage.filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0666)
		if err != nil {
			return err
		}
	}
	go storage.save()
	return nil
}

func (storage *FileStorage) save() {
	for {
		time.Sleep(storage.storeInterval)
		if err := storage.saveToFile(storage.filePath); err != nil {
			storage.log.Debugf("during attempt to store data to file, error occurred: %v", err)
			ctx := context.Background()
			b := retry.WithMaxRetries(3, NewlinearBackoff(time.Second*1))
			err = retry.Do(ctx, b, trySaveFile(storage))
			if err != nil {
				storage.log.Fatalf("unable to load metrics from file - %w", err)
			}
		}
	}
}

func tryLoadFile(filePath string, fs *FileStorage) func(context.Context) error {
	return func(ctx context.Context) error {
		if err := fs.loadFromFile(filePath); err != nil {
			return retry.RetryableError(err)
		}
		return nil
	}
}

func trySaveFile(fs *FileStorage) func(context.Context) error {
	return func(ctx context.Context) error {
		if err := fs.saveToFile(fs.filePath); err != nil {
			return retry.RetryableError(err)
		}
		return nil
	}
}
