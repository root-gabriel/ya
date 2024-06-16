package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/DeneesK/metrics-alerting/internal/models"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sethvargo/go-retry"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var migrationsDir embed.FS

type DBStorage struct {
	log *zap.SugaredLogger
	db  *sql.DB
}

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}

func NewDBStorage(dsn string, log *zap.SugaredLogger) (*DBStorage, error) {
	err := runMigrations(dsn)
	if err != nil {
		return nil, fmt.Errorf("during initializing of new db session, error occurred: %w", err)
	}

	db, err := NewDBSession(dsn)
	if err != nil {
		return nil, fmt.Errorf("during initializing of new db session, error occurred: %w", err)
	}

	return &DBStorage{log: log, db: db}, nil
}

func NewDBSession(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("attempt to establish connection failed - %w", err)
	}
	ctx := context.Background()
	b := retry.WithMaxRetries(3, NewlinearBackoff(time.Second*1))
	err = retry.Do(ctx, b, tryPingDB(db))
	if err != nil {
		return nil, fmt.Errorf("database is not available - %w", err)
	}

	return db, nil
}

func (storage *DBStorage) Ping() error {
	err := storage.db.Ping()
	if err != nil {
		return fmt.Errorf("database is not available - %w", err)
	}
	return nil
}

func (storage *DBStorage) Store(typeMetric string, name string, value interface{}) error {
	switch typeMetric {
	case counterMetric:
		v, ok := value.(int64)
		if !ok {
			return fmt.Errorf("value cannot be cast to integer,  value - %v", value)
		}
		_, err := storage.db.Exec("INSERT INTO metrics VALUES ($1, $2, $3, $4) ON CONFLICT (metricname) DO UPDATE SET counter=metrics.counter+$3", counterMetric, name, v, 0)
		if err != nil {
			return fmt.Errorf("during attempt to store data to database error ocurred: %w", err)
		}
	case gaugeMetric:
		v, ok := value.(float64)
		if !ok {
			return fmt.Errorf("value cannot be cast to integer,  value - %v", value)
		}
		_, err := storage.db.Exec("INSERT INTO metrics VALUES ($1, $2, $3, $4) ON CONFLICT (metricname) DO UPDATE SET gauge=$4", gaugeMetric, name, 0, v)
		if err != nil {
			return fmt.Errorf("insert failed - %w", err)
		}
	default:
		return fmt.Errorf("metric type does not exist, given type: %v", typeMetric)
	}
	return nil
}

func (storage *DBStorage) StoreBatch(metrics []models.Metrics) error {
	ctx := context.Background()
	start := 0
	end := 1000
	for len(metrics) > start {
		if len(metrics)-end < 0 {
			if err := storage.insertBatch(ctx, metrics[start:]); err != nil {
				return fmt.Errorf("postgres db error: %w", err)
			}
			break
		}
		if err := storage.insertBatch(ctx, metrics[start:end]); err != nil {
			return fmt.Errorf("postgres db error: %w", err)
		}
		start += 1000
		end += 1000
	}
	return nil
}

func (storage *DBStorage) GetValue(typeMetric, name string) (Result, bool, error) {
	switch typeMetric {
	case counterMetric:
		var value int64
		row := storage.db.QueryRowContext(context.Background(), "SELECT metrics.counter FROM metrics WHERE metrics.MetricName=$1", name)
		err := row.Scan(&value)
		if err != nil {
			return Result{}, false, fmt.Errorf("query failed - %w", err)
		}
		return Result{Counter: value, Gauge: 0}, true, nil
	case gaugeMetric:
		var value float64
		row := storage.db.QueryRowContext(context.Background(), "SELECT metrics.gauge FROM metrics WHERE metrics.MetricName=$1", name)
		err := row.Scan(&value)
		if err != nil {
			return Result{}, false, fmt.Errorf("query failed - %w", err)
		}
		return Result{Counter: 0, Gauge: value}, true, nil
	default:
		return Result{}, false, fmt.Errorf("metric type does not exist, given type: %v", typeMetric)
	}
}

func (storage *DBStorage) GetCounterMetrics() (map[string]int64, error) {
	metrics := make(map[string]int64, 0)
	var name string
	var v int64
	rows, err := storage.db.QueryContext(context.Background(), "SELECT metrics.metricname, metrics.counter FROM metrics WHERE metrics.metrictype=$1", counterMetric)
	if err != nil {
		return nil, fmt.Errorf("query failed - %w", err)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("row returned error - %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&name, &v)
		if err != nil {
			return nil, fmt.Errorf("during attempt to scan rows error ocurred - %v", err)
		}
		metrics[name] = v
	}
	return metrics, nil
}

func (storage *DBStorage) GetGaugeMetrics() (map[string]float64, error) {
	metrics := make(map[string]float64, 0)
	var name string
	var v float64
	rows, err := storage.db.QueryContext(context.Background(), "SELECT metrics.metricname, metrics.gauge FROM metrics WHERE metrics.metrictype=$1", gaugeMetric)
	if err != nil {
		return nil, fmt.Errorf("query failed - %w", err)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("row returned error - %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&name, &v)
		if err != nil {
			return nil, fmt.Errorf("during attempt to scan rows error ocurred - %w", err)
		}
		metrics[name] = v
	}
	return metrics, nil
}

func (storage *DBStorage) Close() error {
	return storage.db.Close()
}

func (storage *DBStorage) insertBatch(ctx context.Context, metrics []models.Metrics) error {
	tx, err := storage.db.Begin()
	if err != nil {
		return fmt.Errorf("unable to start transaction an error occurred - %w", err)
	}
	defer tx.Rollback()
	values := make(map[string]models.Metrics, len(metrics))
	var valString []string
	var v []interface{}
	for _, m := range metrics {
		if old, ok := values[m.ID]; ok && m.MType == counterMetric {
			newDelta := *m.Delta + *old.Delta
			metric := models.Metrics{ID: m.ID, MType: m.MType, Delta: &newDelta, Value: m.Value}
			values[m.ID] = metric
			continue
		}
		values[m.ID] = m
	}
	i := 0
	for _, m := range values {
		valString = append(valString, fmt.Sprintf("($%d, $%d, $%d, $%d)", i*4+1, i*4+2, i*4+3, i*4+4))
		v = append(v, m.MType)
		v = append(v, m.ID)
		v = append(v, m.Delta)
		v = append(v, m.Value)
		i++
	}
	smt := "INSERT INTO metrics VALUES %s ON CONFLICT (metricname) DO UPDATE SET gauge=EXCLUDED.gauge, counter=metrics.counter+EXCLUDED.counter"
	smt = fmt.Sprintf(smt, strings.Join(valString, ","))
	_, err = tx.ExecContext(ctx, smt, v...)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("insert batch error: %w", err)
	}
	return tx.Commit()
}

func tryPingDB(db *sql.DB) func(context.Context) error {
	return func(ctx context.Context) error {
		if err := db.PingContext(ctx); err != nil {
			return retry.RetryableError(err)
		}
		return nil
	}
}
