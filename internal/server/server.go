package server

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/dtt4h/go-marketplace/internal/config"
	mw "github.com/dtt4h/go-marketplace/internal/server/middleware"
)

type Server struct {
	cfg    *config.Config
	log    *slog.Logger
	db     *pgxpool.Pool
	router *chi.Mux
	http   *http.Server
}

func New(cfg *config.Config, log *slog.Logger, db *pgxpool.Pool) *Server {
	s := &Server{
		cfg:    cfg,
		log:    log,
		db:     db,
		router: chi.NewRouter(),
	}
	s.setupMiddleware()
	s.setupRoutes()
	return s
}

func (s *Server) setupMiddleware() {
	s.router.Use(chimw.RequestID)
	s.router.Use(chimw.RealIP)
	s.router.Use(mw.Logger(s.log))
	s.router.Use(chimw.Recoverer)
	s.router.Use(mw.CORS)
}

func (s *Server) setupRoutes() {
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	s.router.Route("/api/v1", func(r chi.Router) {
	})
}

func (s *Server) Start() error {
	s.http = &http.Server{
		Addr:         ":" + strconv.Itoa(s.cfg.Server.Port),
		Handler:      s.router,
		ReadTimeout:  s.cfg.Server.ReadTimeout,
		WriteTimeout: s.cfg.Server.WriteTimeout,
		IdleTimeout:  s.cfg.Server.IdleTimeout,
	}

	s.log.Info("http server starting", slog.String("addr", s.http.Addr))

	return s.http.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("http server shutting down")
	if s.http != nil {
		return s.http.Shutdown(ctx)
	}
	return nil
}
