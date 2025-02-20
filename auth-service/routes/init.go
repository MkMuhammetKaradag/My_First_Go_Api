package routes

import (
	"net/http"
	"time"

	"github.com/MKMuhammetKaradag/go-microservice/auth-service/controllers"
	"github.com/MKMuhammetKaradag/go-microservice/auth-service/repository"
	"github.com/MKMuhammetKaradag/go-microservice/auth-service/websocket"
	"github.com/MKMuhammetKaradag/go-microservice/shared/messaging"
	"github.com/MKMuhammetKaradag/go-microservice/shared/middlewares"
	"github.com/MKMuhammetKaradag/go-microservice/shared/redisrepo"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
)

func init() {
	// Register metrics with Prometheus
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
}

// PrometheusMiddleware collects metrics for each request
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap ResponseWriter to capture status code
		wrapped := wrapResponseWriter(w)

		next.ServeHTTP(wrapped, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, string(wrapped.status)).Inc()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}

// ResponseWriterWrapper captures the status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func CreateServer(rabbitMQ *messaging.RabbitMQ, sessionRepo *redisrepo.RedisRepository, userRepo *repository.UserRepository) *chi.Mux {
	authController := controllers.NewAuthController(rabbitMQ, sessionRepo)
	authMiddleware := middlewares.NewAuthMiddleware(sessionRepo)
	hub := websocket.NewHub()
	go hub.Run()
	go hub.ListenRedisStatus(sessionRepo)

	wsController := controllers.NewWebSocketController(hub, userRepo, sessionRepo)
	r := chi.NewRouter()

	r.Use(PrometheusMiddleware)
	// r.Handle("/metrics", http.HandlerFunc(promhttp.Handler().ServeHTTP))
	r.Mount("/metrics", promhttp.Handler())
	r.Route("/auth", func(r chi.Router) {
		r.Use(middlewares.Logger)
		r.Post("/signUp", authController.SignUp)
		r.Post("/activationUser", authController.ActivationUser)
		r.Post("/signIn", authController.SignIn)
		r.Post("/forgotPassword", authController.ForgotPassword)
		r.Post("/resetPassword", authController.ResetPassword)

		r.Group(func(protectedRouter chi.Router) {
			protectedRouter.Use(authMiddleware.Authenticate)

			protectedRouter.Post("/logout", authController.Logout)
			protectedRouter.Get("/me", authController.Logout)
			protectedRouter.Get("/protected", controllers.Protected)
			protectedRouter.Post("/updateStatus", authController.UpdateStatus)
			protectedRouter.Get("/ws", wsController.HandleWebSocket)
		})
	})

	// r.Get("/auth/ws", wsController.HandleWebSocket)
	return r
}
