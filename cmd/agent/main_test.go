package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DeneesK/metrics-alerting/internal/models"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
)

func Test_postReport(t *testing.T) {
	v := 10.5
	type args struct {
		metrics []models.Metrics
	}
	tests := []struct {
		name            string
		args            args
		wantContentType string
		wantCode        int
	}{
		{
			name: "positive test #1",
			args: args{
				[]models.Metrics{{ID: "PollCount", MType: "gauge", Value: &v}},
			},
			wantContentType: "application/json",
			wantCode:        200,
		},
	}
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = retryMax
	retryClient.RetryWaitMin = retryWaitMin
	retryClient.RetryWaitMax = retryWaitMax

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					assert.Equal(t, r.Header.Get("Content-Type"), test.wantContentType)
					w.WriteHeader(http.StatusOK)
				}
				w.WriteHeader(http.StatusMethodNotAllowed)
			}))
			statusCode, err := sendBatch(retryClient, ts.URL, test.args.metrics, nil)
			assert.NoError(t, err)
			assert.Equal(t, statusCode, test.wantCode)
			ts.Close()
		})
	}
}
