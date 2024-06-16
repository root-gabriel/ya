package metriccollector

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

const (
	memStatsLen = 17
	cpuStatsLen = 2
)

type Metrics struct {
	mx           sync.Mutex
	stats        runtime.MemStats
	vmstats      mem.VirtualMemoryStat
	cpuUsage     float64
	pollCount    int64
	randomValue  float64
	pollInterval time.Duration
}

type RuntimeMetrics struct {
	memMetrics map[string]uint64
	cpuMetrics map[string]float64
}

func (rm *RuntimeMetrics) GetMemMetrics() map[string]uint64 {
	memMetricsCopy := make(map[string]uint64, len(rm.memMetrics))
	for k, v := range rm.memMetrics {
		memMetricsCopy[k] = v
	}
	return memMetricsCopy
}

func (rm *RuntimeMetrics) GetCPUMetrics() map[string]float64 {
	cpuMetricCopy := make(map[string]float64, len(rm.cpuMetrics))
	for k, v := range rm.cpuMetrics {
		cpuMetricCopy[k] = v
	}
	return cpuMetricCopy
}

func NewCollector(pollInterval int) Metrics {
	return Metrics{pollInterval: time.Duration(pollInterval) * time.Second}
}

func (ms *Metrics) pollMetrics() {
	ms.mx.Lock()
	defer ms.mx.Unlock()
	runtime.ReadMemStats(&ms.stats)
	ms.randomValue = rand.Float64()
	ms.pollCount += 1
}

func (ms *Metrics) StartCollect(ctx context.Context) {
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
			ms.pollMetrics()
			time.Sleep(ms.pollInterval)
		}
	}
}

func (ms *Metrics) pollAddionalMetrics() error {
	ms.mx.Lock()
	defer ms.mx.Unlock()
	vm, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("collecting additional memory metrics failed %w", err)
	}
	ms.vmstats = *vm
	cp, err := cpu.Percent(0, false)
	if err != nil {
		return fmt.Errorf("collecting additional cpu usage metrics failed %w", err)
	}
	ms.cpuUsage = cp[0]
	ms.randomValue = rand.Float64()
	ms.pollCount += 1
	return nil
}

func (ms *Metrics) StartAdditionalCollect(ctx context.Context) {
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
			ms.pollAddionalMetrics()
			time.Sleep(ms.pollInterval)
		}
	}
}

func (ms *Metrics) FillChanel(ctx context.Context, ch chan RuntimeMetrics, reportInterval time.Duration) {
loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
			ch <- ms.GetRuntimeMetrics()
			time.Sleep(reportInterval)
		}
	}
}

func (ms *Metrics) GetRuntimeMetrics() RuntimeMetrics {
	ms.mx.Lock()
	defer ms.mx.Unlock()
	memStats := make(map[string]uint64, memStatsLen)
	memStats["Alloc"] = ms.stats.Alloc
	memStats["BuckHashSys"] = ms.stats.BuckHashSys
	memStats["Frees"] = ms.stats.Frees
	memStats["GCSys"] = ms.stats.GCSys
	memStats["HeapAlloc"] = ms.stats.HeapAlloc
	memStats["HeapIdle"] = ms.stats.HeapIdle
	memStats["HeapInuse"] = ms.stats.HeapInuse
	memStats["HeapObjects"] = ms.stats.HeapObjects
	memStats["HeapReleased"] = ms.stats.HeapReleased
	memStats["HeapSys"] = ms.stats.HeapSys
	memStats["Lookups"] = ms.stats.Lookups
	memStats["MCacheInuse"] = ms.stats.MCacheInuse
	memStats["MCacheSys"] = ms.stats.MCacheSys
	memStats["MSpanInuse"] = ms.stats.MSpanInuse
	memStats["MSpanSys"] = ms.stats.MSpanSys
	memStats["NextGC"] = ms.stats.NextGC
	memStats["Mallocs"] = ms.stats.Mallocs
	memStats["LastGC"] = ms.stats.LastGC
	memStats["OtherSys"] = ms.stats.OtherSys
	memStats["PauseTotalNs"] = ms.stats.PauseTotalNs
	memStats["StackInuse"] = ms.stats.StackInuse
	memStats["StackSys"] = ms.stats.StackSys
	memStats["Sys"] = ms.stats.Sys
	memStats["TotalAlloc"] = ms.stats.TotalAlloc
	memStats["NumForcedGC"] = uint64(ms.stats.NumForcedGC)
	memStats["NumGC"] = uint64(ms.stats.NumGC)
	memStats["TotalMemory"] = ms.vmstats.Total
	memStats["FeeMemory"] = ms.vmstats.Free

	cpuStats := make(map[string]float64, cpuStatsLen)
	cpuStats["GCCPUFraction"] = ms.stats.GCCPUFraction
	cpuStats["CPUutilization1"] = ms.cpuUsage

	return RuntimeMetrics{memMetrics: memStats, cpuMetrics: cpuStats}

}

func (ms *Metrics) GetRandomValue() float64 {
	ms.mx.Lock()
	defer ms.mx.Unlock()
	return ms.randomValue
}

func (ms *Metrics) GetPollCount() int64 {
	ms.mx.Lock()
	defer ms.mx.Unlock()
	count := ms.pollCount
	return count
}

func (ms *Metrics) ResetPollCount() {
	ms.mx.Lock()
	defer ms.mx.Unlock()
	ms.pollCount = 0
}
