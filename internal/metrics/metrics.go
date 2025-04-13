package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HTTPRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Общее количество HTTP-запросов.",
		},
		[]string{"handler", "method", "code"},
	)
	HTTPResponseDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_duration_seconds",
			Help:    "Время ответа HTTP-запросов в секундах.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"handler", "method", "code"},
	)

	CreatedPVZTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "business_created_pvz_total",
			Help: "Количество успешно созданных ПВЗ.",
		},
	)
	CreatedReceptionTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "business_created_reception_total",
			Help: "Количество успешно созданных приемок заказов.",
		},
	)
	AddedProductsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "business_added_products_total",
			Help: "Количество успешно добавленных товаров.",
		},
	)
)

func NewRegistry() *prometheus.Registry {
	reg := prometheus.NewRegistry()
	reg.MustRegister(HTTPRequestTotal, HTTPResponseDuration, CreatedPVZTotal, CreatedReceptionTotal, AddedProductsTotal)
	return reg
}

func Handler() http.Handler {
	return promhttp.HandlerFor(NewRegistry(), promhttp.HandlerOpts{})
}
