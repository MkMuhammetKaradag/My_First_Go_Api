package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/MKMuhammetKaradag/go-microservice/chat-service/controllers"
	"github.com/MKMuhammetKaradag/go-microservice/chat-service/repository"
	"github.com/MKMuhammetKaradag/go-microservice/chat-service/websocket"
	"github.com/MKMuhammetKaradag/go-microservice/shared/messaging"
	"github.com/MKMuhammetKaradag/go-microservice/shared/middlewares"
	"github.com/MKMuhammetKaradag/go-microservice/shared/redisrepo"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
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
func CreateServer(rabbitMQ *messaging.RabbitMQ, chatRepo *repository.ChatRepository, sessionRepo *redisrepo.RedisRepository) *chi.Mux {
	chatController := controllers.NewChatController(rabbitMQ, sessionRepo)
	authMiddleware := middlewares.NewAuthMiddleware(sessionRepo)
	hub := websocket.NewHub()
	go hub.Run()
	go hub.ListenRedisSendMessage(sessionRepo)
	wsController := controllers.NewWebSocketController(hub, chatRepo, sessionRepo)
	r := chi.NewRouter()
	r.Use(middlewares.Logger)
	r.Use(PrometheusMiddleware)
	r.Mount("/metrics", promhttp.Handler())
	r.Route("/chat", func(r chi.Router) {
		r.Get("/chat", func(w http.ResponseWriter, r *http.Request) {

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"message": "get Chat",
				"chat":    "chat",
			})

		})
		// Swagger UI'yi "/swagger/" endpointine bağla
		r.Get("/swagger/*", httpSwagger.WrapHandler)
		r.Group(func(protectedRouter chi.Router) {
			protectedRouter.Use(authMiddleware.Authenticate)
			protectedRouter.Post("/create", chatController.CreateChat)
			protectedRouter.Get("/{chatID}", chatController.CreateChat)
			protectedRouter.Get("/myChats", chatController.GetMyChats)
			protectedRouter.Post("/message/create", chatController.SendMessage)
			protectedRouter.Post("/addParticipants", chatController.AddParticipants)
			protectedRouter.Post("/removeParticipants", chatController.RemoveParticipants)
			protectedRouter.Post("/leave/{chatID}", chatController.LeaveChat)
			protectedRouter.Get("/chatDetail", chatController.GetChatUsers)
			protectedRouter.Get("/chatlisten/{chatID}", wsController.HandleWebSocket)
			protectedRouter.Get("/messages", chatController.GetChatMessages)
		})
	})

	return r
}
