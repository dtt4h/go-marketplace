package server

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/dtt4h/go-marketplace/internal/cache"
	"github.com/dtt4h/go-marketplace/internal/config"
	"github.com/dtt4h/go-marketplace/internal/metrics"
	mw "github.com/dtt4h/go-marketplace/internal/server/middleware"
	"github.com/dtt4h/go-marketplace/internal/storage"
)

type Server struct {
	cfg     *config.Config
	log     *slog.Logger
	db      *pgxpool.Pool
	rdb     *redis.Client
	store   storage.ObjectStorage
	router  *chi.Mux
	http    *http.Server
	cache   *cache.CatalogCache
	metrics *metrics.Registry
}

func New(cfg *config.Config, log *slog.Logger, db *pgxpool.Pool, store storage.ObjectStorage, rdb *redis.Client) *Server {
	s := &Server{
		cfg:     cfg,
		log:     log,
		db:      db,
		rdb:     rdb,
		store:   store,
		router:  chi.NewRouter(),
		metrics: metrics.New(),
		cache:   cache.NewCatalogCache(rdb, log, 60*time.Second),
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
	s.router.Use(mw.Security(s.cfg))
	s.router.Use(mw.CORS(s.cfg))
	s.router.Use(s.metrics.Middleware)
	s.router.Use(mw.MaxBodySize(10 << 20)) // 10MB default
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

func (s *Server) healthLive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) healthReady(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	checks := map[string]string{}
	allOK := true

	// DB ping
	if err := s.db.Ping(ctx); err != nil {
		checks["database"] = "fail"
		allOK = false
	} else {
		checks["database"] = "ok"
	}

	// Redis ping (optional — app works without it in dev mode)
	if s.rdb != nil {
		if err := s.rdb.Ping(ctx).Err(); err != nil {
			checks["redis"] = "fail"
			allOK = false
		} else {
			checks["redis"] = "ok"
		}
	} else {
		checks["redis"] = "skipped"
	}

	w.Header().Set("Content-Type", "application/json")
	if !allOK {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	// Manual JSON to avoid extra encoding import
	body := `{"status":"ok","checks":{"database":"` + checks["database"] + `","redis":"` + checks["redis"] + `"}}`
	if !allOK {
		body = `{"status":"fail","checks":{"database":"` + checks["database"] + `","redis":"` + checks["redis"] + `"}}`
	}
	_, _ = w.Write([]byte(body))
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("http server shutting down")
	if s.http != nil {
		return s.http.Shutdown(ctx)
	}
	return nil
}
