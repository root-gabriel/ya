package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/lionslon/go-yapmetrics/internal/config"
	"github.com/lionslon/go-yapmetrics/internal/models"
	"go.uber.org/zap"
	"math/rand"
	"runtime"
	"time"
)

var valuesGauge = map[string]float64{}
var pollCount uint64

func main() {

	cfg := config.NewClient()

	pollTicker := time.NewTicker(time.Duration(cfg.PollInterval) * time.Second)
	defer pollTicker.Stop()
	reportTicker := time.NewTicker(time.Duration(cfg.ReportInterval) * time.Second)
	defer reportTicker.Stop()

	for {
		select {
		case <-pollTicker.C:
			getMetrics()
		case <-reportTicker.C:
			postQueries(cfg)
		}
	}
}

func getMetrics() {
	var rtm runtime.MemStats

	pollCount += 1
	runtime.ReadMemStats(&rtm)

	valuesGauge["Alloc"] = float64(rtm.Alloc)
	valuesGauge["BuckHashSys"] = float64(rtm.BuckHashSys)
	valuesGauge["Frees"] = float64(rtm.Frees)
	valuesGauge["GCCPUFraction"] = float64(rtm.GCCPUFraction)
	valuesGauge["HeapAlloc"] = float64(rtm.HeapAlloc)
	valuesGauge["HeapIdle"] = float64(rtm.HeapIdle)
	valuesGauge["HeapInuse"] = float64(rtm.HeapInuse)
	valuesGauge["HeapObjects"] = float64(rtm.HeapObjects)
	valuesGauge["HeapReleased"] = float64(rtm.HeapReleased)
	valuesGauge["HeapSys"] = float64(rtm.HeapSys)
	valuesGauge["LastGC"] = float64(rtm.LastGC)
	valuesGauge["Lookups"] = float64(rtm.Lookups)
	valuesGauge["MCacheInuse"] = float64(rtm.MCacheInuse)
	valuesGauge["MCacheSys"] = float64(rtm.MCacheSys)
	valuesGauge["MSpanInuse"] = float64(rtm.MSpanInuse)
	valuesGauge["MSpanSys"] = float64(rtm.MSpanSys)
	valuesGauge["Mallocs"] = float64(rtm.Mallocs)
	valuesGauge["NextGC"] = float64(rtm.NextGC)
	valuesGauge["NumForcedGC"] = float64(rtm.NumForcedGC)
	valuesGauge["NumGC"] = float64(rtm.NumGC)
	valuesGauge["OtherSys"] = float64(rtm.OtherSys)
	valuesGauge["PauseTotalNs"] = float64(rtm.PauseTotalNs)
	valuesGauge["StackInuse"] = float64(rtm.StackInuse)
	valuesGauge["StackSys"] = float64(rtm.StackSys)
	valuesGauge["Sys"] = float64(rtm.Sys)
	valuesGauge["TotalAlloc"] = float64(rtm.TotalAlloc)
}

func postQueries(cfg *config.ClientConfig) {
	url := fmt.Sprintf("http://%s/update/", cfg.Addr)
	client := retryablehttp.NewClient()
	client.RetryMax = 3
	client.RetryWaitMin = time.Second * 1
	client.RetryWaitMax = time.Second * 5

	for k, v := range valuesGauge {
		postJSON(client, url, models.Metrics{ID: k, MType: "gauge", Value: &v})
	}
	pc := int64(pollCount)
	postJSON(client, url, models.Metrics{ID: "PollCount", MType: "counter", Delta: &pc})
	r := rand.Float64()
	postJSON(client, url, models.Metrics{ID: "RandomValue", MType: "gauge", Value: &r})
	pollCount = 0
}

func postJSON(c *retryablehttp.Client, url string, m models.Metrics) {
	js, err := json.Marshal(m)
	if err != nil {
		zap.S().Error(err)
	}

	gz, err := compress(js)
	if err != nil {
		zap.S().Error(err)
	}

	req, err := retryablehttp.NewRequest("POST", url, gz)
	if err != nil {
		zap.S().Error(err)
	}

	req.Header.Add("content-type", "application/json")
	req.Header.Add("content-encoding", "gzip")
	resp, err := c.Do(req)
	if err != nil {
		zap.S().Error(err)
	}
	defer resp.Body.Close()
}

func compress(b []byte) ([]byte, error) {
	var bf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&bf, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}
	_, err = gz.Write(b)
	if err != nil {
		return nil, err
	}
	gz.Close()
	return bf.Bytes(), nil
}
