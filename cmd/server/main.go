package main

import (
	"log"
	"net/http"
	"strconv"

	authClerk "devarminas/kcal-track-api/internal/auth/clerk"
	"devarminas/kcal-track-api/internal/cache"
	kcalMiddleware "devarminas/kcal-track-api/internal/middleware"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

func NewJWKSClient(secretKey string) *jwks.Client {
	cfg := &clerk.ClientConfig{
		BackendConfig: clerk.BackendConfig{
			Key: clerk.String(secretKey),
		},
	}
	return jwks.NewClient(cfg)
}

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	config, err := NewConfigFromEnv()
	if err != nil {
		logger.Panic("failed to read env variables", zap.Error(err))
	}

	store := cache.NewInMemoryCache()
	client := NewJWKSClient(config.ClerkSecret)
	verifier := authClerk.NewVerifier(store, client, logger)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(kcalMiddleware.ZapLogger(logger))
	r.Use(middleware.Recoverer)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	r.With(kcalMiddleware.ClerkAuth(verifier, logger)).Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	log.Printf("listening on :%v", config.Port)
	if err := http.ListenAndServe(":"+strconv.Itoa(config.Port), r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
