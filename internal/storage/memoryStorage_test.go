package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateCounter(t *testing.T) {
	s := NewMem()
	testCases := []struct {
		name        string
		metricsName string
		value       int64
		want        int64
	}{
		{name: "UpdateCounter() #1", metricsName: "testCounter1", value: 10, want: 10},
		{name: "UpdateCounter() #2", metricsName: "testCounter2", value: 1, want: 1},
		{name: "UpdateCounter() #3", metricsName: "testCounter1", value: 1, want: 11},
		{name: "UpdateCounter() #4", metricsName: "testCounter1", value: 10000, want: 10011},
		{name: "UpdateCounter() #5", metricsName: "testCounter2", value: 0, want: 1},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			s.UpdateCounter(test.metricsName, test.value)
			assert.Equal(t, counter(test.want), s.CounterData[test.metricsName])
		})
	}
}

func TestUpdateGauge(t *testing.T) {
	s := NewMem()
	testCases := []struct {
		name        string
		metricsName string
		value       float64
		want        float64
	}{
		{name: "UpdateGauge() #1", metricsName: "testGauge1", value: 1, want: 1.0},
		{name: "UpdateGauge() #2", metricsName: "testGauge2", value: 1.0, want: 1.0},
		{name: "UpdateGauge() #3", metricsName: "testGauge1", value: 10000, want: 10000.0},
		{name: "UpdateGauge() #4", metricsName: "testGauge2", value: 0, want: 0.0},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			s.UpdateGauge(test.metricsName, test.value)
			assert.Equal(t, gauge(test.want), s.GaugeData[test.metricsName])
		})
	}
}
