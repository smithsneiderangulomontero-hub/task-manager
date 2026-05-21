package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// SecurityHeaders agrega headers de seguridad HTTP
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'none'")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store, max-age=0")
		next.ServeHTTP(w, r)
	})
}

// RateLimit: máximo de requests por segundo
type rateLimiter struct {
	tokens  chan struct{}
	closeCh chan struct{}
}

func newRateLimiter(rps int) *rateLimiter {
	rl := &rateLimiter{
		tokens:  make(chan struct{}, rps),
		closeCh: make(chan struct{}),
	}
	for range rps {
		rl.tokens <- struct{}{}
	}
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				for range rps {
					select {
					case rl.tokens <- struct{}{}:
					default:

					}
				}
			case <-rl.closeCh:
				return
			}
		}
	}()
	return rl
}

func (rl *rateLimiter) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-rl.tokens:
			next.ServeHTTP(w, r)
		default:
			slog.Warn("rate limit exceeded", "ip", r.RemoteAddr, "path", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate limit exceeded"}`))
		}
	})
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(securityHeaders)

	// 100 requests por segundo
	rl := newRateLimiter(100)
	r.Use(rl.middleware)

	r.Get("/health", h.HealthCheck)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/tasks", func(r chi.Router) {
			r.Get("/", h.ListTasks)
			r.Post("/", h.CreateTask)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", h.GetTask)
				r.Put("/", h.UpdateTask)
				r.Delete("/", h.DeleteTask)
			})
		})
	})
}
