package api

import (
	"github.com/labstack/echo/v4"
	"github.com/lionslon/go-yapmetrics/internal/config"
	"github.com/lionslon/go-yapmetrics/internal/handlers"
	"github.com/lionslon/go-yapmetrics/internal/middlewares"
	"github.com/lionslon/go-yapmetrics/internal/storage"
	"go.uber.org/zap"
)

type APIServer struct {
	Cfg             *config.ServerConfig
	echo            *echo.Echo
	st              *storage.MemStorage
	storageProvider storage.StorageWorker
}

func New() *APIServer {
	apiS := &APIServer{}
	cfg := config.NewServer()
	apiS.Cfg = cfg
	apiS.echo = echo.New()
	apiS.st = storage.NewMem()

	var err error
	switch cfg.GetProvider() {
	case storage.FileProvider:
		apiS.storageProvider = storage.NewFileProvider(cfg.FilePath, cfg.StoreInterval, apiS.st)
	case storage.DBProvider:
		apiS.storageProvider, err = storage.NewDBProvider(cfg.DatabaseDSN, cfg.StoreInterval, apiS.st)
	}
	if err != nil {
		zap.S().Error(err)
	}
	if cfg.Restore {
		err := apiS.storageProvider.Restore()
		if err != nil {
			zap.S().Error(err)
		}
	}

	if cfg.StoreIntervalNotZero() {
		go apiS.storageProvider.IntervalDump()
	}

	apiS.echo.Use(middlewares.WithLogging())
	apiS.echo.Use(middlewares.GzipUnpacking())
	if cfg.SignPass != "" {
		apiS.echo.Use(middlewares.CheckSignReq(cfg.SignPass))
	}

	return apiS
}

func (a *APIServer) RegisterRoutes(e *echo.Echo) {
	handler := handlers.New(a.st)
	e.GET("/", handler.AllMetricsValues())
	e.POST("/value/", handler.GetValueJSON())
	e.GET("/value/:typeM/:nameM", handler.MetricsValue())
	e.POST("/update/", handler.UpdateJSON())
	e.POST("/update/:typeM/:nameM/:valueM", handler.UpdateMetrics())
	e.POST("/updates/", handler.UpdatesJSON())
	e.GET("/ping", handler.PingDB(a.storageProvider))
}

func (a *APIServer) Start() error {
	err := a.echo.Start(a.Cfg.Addr)
	if err != nil {
		zap.S().Fatal(err)
	}

	return nil
}

