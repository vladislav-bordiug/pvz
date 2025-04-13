package app

import (
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"pvz/internal/database"
	"pvz/internal/metrics"
	"pvz/internal/middleware"
	pb "pvz/internal/pb/pvz_v1"
	"pvz/internal/services"
	grpch "pvz/internal/transport/grpc"
	"pvz/internal/transport/rest"
)

type App struct {
	pool           database.DBPool
	port           string
	grpcport       string
	prometheusport string
	jwtSecret      []byte
}

func NewApp(pool database.DBPool, port string, grpcport string, prometheusport string, jwtSecret []byte) *App {
	return &App{pool: pool, port: port, grpcport: grpcport, prometheusport: prometheusport, jwtSecret: jwtSecret}
}

func (a *App) Run() error {
	logrus.Info("Приложение запускается...")

	db := database.NewPGXDatabase(a.pool)
	logrus.Info("Соединение с базой данных установлено")

	service := services.NewService(db, a.jwtSecret)
	logrus.Info("Сервис инициализирован")

	handler := rest.NewHandler(service)
	logrus.Info("Обработчики REST-запросов инициализированы")

	middle := middleware.NewMiddleware(a.jwtSecret)

	router := mux.NewRouter()
	router.Use(middle.MetricsMiddleware)

	router.HandleFunc("/dummyLogin", handler.DummyLoginHandler).Methods("POST")
	router.HandleFunc("/register", handler.RegisterHandler).Methods("POST")
	router.HandleFunc("/login", handler.LoginHandler).Methods("POST")

	api := router.PathPrefix("/").Subrouter()
	api.Use(middle.AuthMiddleware)

	api.HandleFunc("/pvz", handler.CreatePVZHandler).Methods("POST")
	api.HandleFunc("/pvz", handler.ListPVZHandler).Methods("GET")
	api.HandleFunc("/pvz/{pvzId}/close_last_reception", handler.CloseLastReceptionHandler).Methods("POST")
	api.HandleFunc("/pvz/{pvzId}/delete_last_product", handler.DeleteLastProductHandler).Methods("POST")

	api.HandleFunc("/receptions", handler.CreateReceptionHandler).Methods("POST")
	api.HandleFunc("/products", handler.AddProductHandler).Methods("POST")
	logrus.Info("Маршруты зарегистрированы")

	const readmax, writemax, idlemax = 5 * time.Second, 10 * time.Second, 120 * time.Second
	server := &http.Server{
		Addr:         ":" + a.port,
		Handler:      router,
		ReadTimeout:  readmax,
		WriteTimeout: writemax,
		IdleTimeout:  idlemax,
	}
	logrus.Infof("HTTP сервер настроен и будет запущен на порту: %s", a.port)

	lis, err := net.Listen("tcp", ":"+a.grpcport)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()

	srv := grpch.NewGrpcServer(service)
	pb.RegisterPVZServiceServer(grpcServer, srv)
	reflection.Register(grpcServer)
	logrus.Infof("GRPC сервер запущен на порту: %s", a.grpcport)

	errChan := make(chan error, 3)

	go func() {
		logrus.Info("GRPC сервер начинает обслуживание подключений")
		errChan <- grpcServer.Serve(lis)
	}()

	go func() {
		logrus.Info("HTTP сервер начинает обслуживание запросов")
		errChan <- server.ListenAndServe()
	}()

	go func() {
		logrus.Infof("Сервер метрик запущен на порту: %s", a.prometheusport)
		http.Handle("/metrics", metrics.Handler())
		errChan <- http.ListenAndServe(":"+a.prometheusport, nil)
	}()

	err = <-errChan
	return err
}
