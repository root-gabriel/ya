package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DeneesK/metrics-alerting/internal/api"
	"github.com/DeneesK/metrics-alerting/internal/logger"
	"github.com/DeneesK/metrics-alerting/internal/models"
	"github.com/DeneesK/metrics-alerting/internal/storage"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func routerWithoutMiddlewares(ms storage.Storage, logging *zap.SugaredLogger) chi.Router {
	r := chi.NewRouter()
	r.Post("/update/{metricType}/{metricName}/{value}", api.Update(ms, logging))
	r.Get("/value/{metricType}/{metricName}", api.Value(ms, logging))
	r.Post("/update/", api.UpdateJSON(ms, logging))
	r.Post("/value/", api.ValueJSON(ms, logging))
	r.Get("/", api.Metrics(ms, logging))
	return r
}

func Test_update_json(t *testing.T) {
	var metricCounter int64 = 1
	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name string
		args models.Metrics
		want want
	}{
		{
			name: "positive test #1",
			args: models.Metrics{ID: "metric", MType: "counter", Delta: &metricCounter},
			want: want{
				code:        200,
				contentType: "application/json",
			},
		},
	}
	log, err := logger.LoggerInitializer("fatal")
	if err != nil {
		t.Error(err)
		return
	}
	ms, err := storage.NewStorage("", 0, false, log, "")
	if err != nil {
		t.Error(err)
		return
	}
	ts := httptest.NewServer(routerWithoutMiddlewares(ms, log))
	defer ts.Close()
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			res, err := json.Marshal(&test.args)
			require.NoError(t, err)
			buf := bytes.NewBuffer(res)
			request, err := http.NewRequest(http.MethodPost, ts.URL+"/update/", buf)
			require.NoError(t, err)
			resp, err := ts.Client().Do(request)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, resp.StatusCode, test.want.code)
			assert.Equal(t, resp.Header.Get("Content-Type"), test.want.contentType)
		})
	}
}

func Test_update(t *testing.T) {
	type want struct {
		code        int
		contentType string
	}
	tests := []struct {
		name string
		args string
		want want
	}{
		{
			name: "positive test #1",
			args: "/update/counter/metric/1",
			want: want{
				code:        200,
				contentType: "text/plain",
			},
		},
		{
			name: "negative test: wrong value #1",
			args: "/update/counter/metric/b",
			want: want{
				code:        400,
				contentType: "",
			},
		},
		{
			name: "negative test: missing metric name #2",
			args: "/update/counter/",
			want: want{
				code:        404,
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	log, err := logger.LoggerInitializer("fatal")
	if err != nil {
		t.Error(err)
		return
	}
	ms, err := storage.NewStorage("", 0, false, log, "")
	if err != nil {
		t.Error(err)
		return
	}
	ts := httptest.NewServer(routerWithoutMiddlewares(ms, log))
	defer ts.Close()
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodPost, ts.URL+test.args, nil)
			require.NoError(t, err)
			resp, err := ts.Client().Do(request)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, resp.StatusCode, test.want.code)
			assert.Equal(t, resp.Header.Get("Content-Type"), test.want.contentType)
		})
	}
}
