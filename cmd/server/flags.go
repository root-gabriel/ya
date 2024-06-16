package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Conf struct {
	runAddr       string
	logLevel      string
	storeInterval int
	filePath      string
	isRestore     bool
	dsn           string
	hashKey       []byte
}

func parseFlags() (Conf, error) {
	var conf Conf
	var hashString string
	flag.StringVar(&conf.runAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&conf.logLevel, "l", "debug", "log level")
	flag.StringVar(&conf.filePath, "f", "tmp/metrics-db.json", "path to store file")
	flag.StringVar(&conf.dsn, "d", "", "database's dsn connection configs")
	flag.StringVar(&hashString, "k", "", "the key to calculate hash")
	flag.BoolVar(&conf.isRestore, "r", true, "load saved data")
	flag.IntVar(&conf.storeInterval, "i", 5, "interval of storing data on disk")
	flag.Parse()
	if envRunAddr, ok := os.LookupEnv("ADDRESS"); ok {
		conf.runAddr = envRunAddr
	}
	if hashString != "" {
		conf.hashKey = []byte(hashString)
	}
	if envHashKey, ok := os.LookupEnv("KEY"); ok {
		conf.hashKey = []byte(envHashKey)
	}
	if envLogLVL, ok := os.LookupEnv("LOG_LEVEL"); ok {
		conf.logLevel = envLogLVL
	}
	if envFilePath, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		conf.filePath = envFilePath
	}
	if envDBDSN, ok := os.LookupEnv("DATABASE_DSN"); ok {
		conf.dsn = envDBDSN
	}
	if envIsRestore, ok := os.LookupEnv("RESTORE"); ok {
		envIsRestore, err := strconv.ParseBool(envIsRestore)
		if err != nil {
			return Conf{}, fmt.Errorf("value of RESTORE is not a boolean: %w", err)
		}
		conf.isRestore = envIsRestore
	}
	if envStoreInterval, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		fsi, err := strconv.Atoi(envStoreInterval)
		if err != nil {
			return Conf{}, fmt.Errorf("value of STORE_INTERVAL is not a integer: %w", err)
		}
		conf.storeInterval = fsi
	}
	return conf, nil
}
