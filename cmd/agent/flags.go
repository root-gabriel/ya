package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type Conf struct {
	runAddr           string
	hashKey           []byte
	reportingInterval int
	pollingInterval   int
	rateLimit         int
}

func parseFlags() (Conf, error) {
	var hashString string
	var conf Conf
	flag.StringVar(&conf.runAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&hashString, "k", "", "the key to calculate hash")
	flag.IntVar(&conf.reportingInterval, "r", 10, "interval of sending metrics to the server")
	flag.IntVar(&conf.pollingInterval, "p", 2, "interval of polling metrics from the runtime package")
	flag.IntVar(&conf.rateLimit, "l", 1, "number of outgoing requests to the server at the same time")
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

	if envreportInterval, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		fri, err := strconv.Atoi(envreportInterval)
		if err != nil {
			return Conf{}, fmt.Errorf("value of REPORT_INTERVAL is not a integer: %w", err)
		}
		conf.reportingInterval = fri
	}
	if envrateLimit, ok := os.LookupEnv("RATE_LIMIT"); ok {
		rl, err := strconv.Atoi(envrateLimit)
		if err != nil {
			return Conf{}, fmt.Errorf("value of RATE_LIMIT is not a integer: %w", err)
		}
		conf.rateLimit = rl
	}
	if envpollInterval, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		fpi, err := strconv.Atoi(envpollInterval)
		if err != nil {
			return Conf{}, fmt.Errorf("value of POLL_INTERVAL is not a integer: %w", err)
		}
		conf.pollingInterval = fpi
	}
	return conf, nil
}
