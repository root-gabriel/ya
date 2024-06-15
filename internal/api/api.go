package api

import (
	"github.com/labstack/echo/v4"
	"github.com/root-gabriel/ya/internal/config"
	"github.com/root-gabriel/ya/internal/handlers"
	"github.com/root-gabriel/ya/internal/middlewares"
	"github.com/root-gabriel/ya/internal/storage"
	"go.uber.org/zap"
	"log"
)

type APIServer struct {
	Cfg             *config.ServerConfig
	Echo            *echo.Echo
	Storage         *storage.MemStorage
	StorageProvider storage.StorageWorker
}

func New() *APIServer {
	apiS := &APIServer{}
	cfg := config.NewServer()
	apiS.Cfg = cfg
	apiS.Echo = echo.New()
	apiS.Storage = storage.NewMem()

	handler := handlers.New(apiS.Storage)
	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)
	defer logger.Sync()

	var storageProvider storage.StorageWorker
	var err error
	switch cfg.GetProvider() {
	case storage.FileProvider:
		storageProvider = storage.NewFileProvider(cfg.FilePath, cfg.StoreInterval, apiS.Storage)
	case storage.DBProvider:
		storageProvider, err = storage.NewDBProvider(cfg.DatabaseDSN, cfg.StoreInterval, apiS.Storage)
	}
	if err != nil {
		zap.S().Error(err)
	}
	if cfg.Restore {
		err := storageProvider.Restore()
		if err != nil {
			zap.S().Error(err)
		}
	}

	if cfg.StoreIntervalNotZero() {
		go storageProvider.IntervalDump()
	}

	apiS.Echo.Use(middlewares.WithLogging())
	apiS.Echo.Use(middlewares.GzipUnpacking())
	if cfg.SignPass != "" {
		apiS.Echo.Use(middlewares.CheckSignReq(cfg.SignPass))
	}

	apiS.Echo.GET("/", handler.AllMetricsValues())
	apiS.Echo.POST("/value/", handler.GetValueJSON())
	apiS.Echo.GET("/value/:typeM/:nameM", handler.MetricsValue())
	apiS.Echo.POST("/update/", handler.UpdateJSON())
	apiS.Echo.POST("/update/:typeM/:nameM/:valueM", handler.UpdateMetrics())
	apiS.Echo.POST("/updates/", handler.UpdatesJSON())
	apiS.Echo.GET("/ping", handler.PingDB(storageProvider))

	apiS.StorageProvider = storageProvider

	return apiS
}

func (a *APIServer) Start() error {
	err := a.Echo.Start(a.Cfg.Addr)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

