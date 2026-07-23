package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/internal/modules/auth"
	"github.com/dtt4h/go-marketplace/internal/modules/users"
	mw "github.com/dtt4h/go-marketplace/internal/server/middleware"
)

func (s *Server) setupRoutes() {
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	s.router.Route("/api/v1", func(r chi.Router) {
		s.registerAuthRoutes(r)
		s.registerUserRoutes(r)
	})
}

func (s *Server) registerAuthRoutes(r chi.Router) {
	queries := db.New(s.db)

	authRepo := auth.NewAuthRepository(queries)
	authSvc := auth.NewAuthService(authRepo, s.cfg)
	authHandler := auth.NewAuthHandler(authSvc)

	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login", authHandler.Login)
	r.Post("/auth/refresh", authHandler.Refresh)
	r.With(mw.JWTAuth(s.cfg)).Post("/auth/logout", authHandler.Logout)
}

func (s *Server) registerUserRoutes(r chi.Router) {
	queries := db.New(s.db)

	userRepo := users.NewUserRepository(queries)
	userSvc := users.NewUserService(userRepo)
	userHandler := users.NewUserHandler(userSvc)

	r.With(mw.JWTAuth(s.cfg)).Get("/users/me", userHandler.GetProfile)
	r.With(mw.JWTAuth(s.cfg)).Patch("/users/me", userHandler.UpdateProfile)
	r.With(mw.JWTAuth(s.cfg)).Post("/users/me/store", userHandler.CreateStore)
}
