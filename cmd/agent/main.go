package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/DeneesK/metrics-alerting/internal/services/metriccollector"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	conf, err := parseFlags()

	if err != nil {
		return fmt.Errorf("args parse failed %w", err)
	}

	ms := metriccollector.NewCollector(conf.pollingInterval)

	ch := make(chan metriccollector.RuntimeMetrics, conf.rateLimit)
	reportInterval := time.Duration(conf.reportingInterval) * time.Second
	ctx, cancelContext := context.WithCancel(context.Background())

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("client started sending data on %s", conf.runAddr)
	var wg sync.WaitGroup

	collect(ctx, &wg, ch, reportInterval, &ms)

	for i := 0; i < conf.rateLimit; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
		loop:
			for {
				select {
				case metrics := <-ch:
					if err := sendMetrics(metrics, &ms, conf.runAddr, conf.hashKey); err != nil {
						log.Println(err)
					}
				case <-ctx.Done():
					break loop
				}
			}
		}()
	}
	<-termChan
	cancelContext()
	wg.Wait()
	log.Println("All workers were shuted down")
	return nil
}

func collect(ctx context.Context,
	wg *sync.WaitGroup,
	ch chan metriccollector.RuntimeMetrics,
	reportInterval time.Duration,
	ms *metriccollector.Metrics) {

	go mainMetrics(ctx, wg, ms)
	go additionalMetrics(ctx, wg, ms)
	go fillChanel(ctx, wg, ms, ch, reportInterval)

}

func mainMetrics(ctx context.Context, wg *sync.WaitGroup, ms *metriccollector.Metrics) {
	wg.Add(1)
	defer wg.Done()
	ms.StartCollect(ctx)
}

func additionalMetrics(ctx context.Context, wg *sync.WaitGroup, ms *metriccollector.Metrics) {
	wg.Add(1)
	defer wg.Done()
	ms.StartAdditionalCollect(ctx)
}

func fillChanel(ctx context.Context, wg *sync.WaitGroup, ms *metriccollector.Metrics,
	ch chan metriccollector.RuntimeMetrics, reportInterval time.Duration) {
	wg.Add(1)
	defer wg.Done()
	ms.FillChanel(ctx, ch, reportInterval)
}
