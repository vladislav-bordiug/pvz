package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"

	"pvz/internal/metrics"
)

func TestMetricsMiddleware(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("OK"))
	})

	mw := &Middleware{}

	wrappedHandler := mw.MetricsMiddleware(dummyHandler)

	req := httptest.NewRequest(http.MethodGet, "/dummyTest", nil)
	rr := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code, "Ожидался статус 201 Created")
	assert.Equal(t, "OK", rr.Body.String(), "Ответное тело должно быть 'OK'")

	codeStr := strconv.Itoa(rr.Code)

	count := testutil.ToFloat64(metrics.HTTPRequestTotal.WithLabelValues("/dummyTest", "GET", codeStr))
	assert.Equal(t, 1.0, count, "HTTPRequestTotal для /dummyTest должен быть равен 1")

	m, ok := metrics.HTTPResponseDuration.WithLabelValues("/dummyTest", "GET", codeStr).(prometheus.Metric)
	assert.True(t, ok, "Невозможно привести histogram к prometheus.Metric")

	var histMetric dto.Metric
	err := m.Write(&histMetric)
	assert.NoError(t, err)

	hist := histMetric.GetHistogram()
	assert.Equal(t, uint64(1), hist.GetSampleCount(), "Количество наблюдений в гистограмме должно быть 1")
}
