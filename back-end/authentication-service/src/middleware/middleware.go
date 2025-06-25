package middleware

import (
	"authentication-service/src/config"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

type Middleware struct {
	Config *config.Config
}

// NewMiddleware creates a new instance of Middleware
func NewMiddleware(cfg *config.Config) *Middleware {
	return &Middleware{
		Config: cfg,
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func (m *Middleware) RequestTimingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()

			// Wrap response writer to capture status
			recorder := &statusRecorder{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			// Get client IP
			ip := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				ip = forwarded
			}

			// Get user agent
			userAgent := r.UserAgent()

			// Read body safely (if needed for debug)
			var body string
			if r.Body != nil && r.ContentLength > 0 {
				bodyBytes, err := io.ReadAll(r.Body)
				if err == nil {
					body = string(bodyBytes)
					// Reassign the Body since we already read it
					r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			}

			// Call actual handler
			next.ServeHTTP(recorder, r)

			duration := time.Since(startTime)

			// Trim and limit body log size
			bodyStr := strings.TrimSpace(body)
			if len(bodyStr) > 500 {
				bodyStr = bodyStr[:500] + " ...(truncated)"
			}

			// Create formatted log message
			logMsg := fmt.Sprintf(
				"[%d] %s %s (%v) | IP: %s | UA: %s | Body: %s",
				recorder.status,
				r.Method,
				r.URL.Path,
				duration,
				ip,
				userAgent,
				bodyStr,
			)

			m.Config.Logger.Info(logMsg)
		})
	}
}

func (m *Middleware) RateLimiterMiddleware() func(http.Handler) http.Handler {
	// Limit each IP to 100 requests per 1 minute (can tune per endpoint later)
	return httprate.LimitByIP(100, 1*time.Minute)
}

// SetupMiddleware sets up all global middleware
func (m *Middleware) SetupMiddleware(mux *chi.Mux) {
	mux.Use(m.CORSMiddleware())
	mux.Use(middleware.Heartbeat("/ping"))
	mux.Use(middleware.Recoverer)

	// Add this timing + status middleware
	mux.Use(m.RequestTimingMiddleware())

	mux.Use(m.RateLimiterMiddleware())
	m.Config.Logger.Info("✅ Middleware initialized successfully")
}

// CORSMiddleware returns a cors.Handler middleware
func (m *Middleware) CORSMiddleware() func(http.Handler) http.Handler {
	corsOrigins := os.Getenv("BROKER_SERVICE_CORS_ALLOWED_ORIGINS")
	allowedOrigins := strings.Split(corsOrigins, ",")

	// Log CORS setup initialization
	m.Config.Logger.Info("✅ CORS middleware initialized with allowed origins: " + strings.Join(allowedOrigins, ", "))

	return cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
