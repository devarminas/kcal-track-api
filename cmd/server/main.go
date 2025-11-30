package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"devarminas/kcal-track-api/ent"
	"devarminas/kcal-track-api/ent/product"
	authClerk "devarminas/kcal-track-api/internal/auth/clerk"
	"devarminas/kcal-track-api/internal/cache"
	kcalMiddleware "devarminas/kcal-track-api/internal/middleware"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
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

	// TODO: remove
	// connectionString := "host=localhost port=<port> user=fly-user dbname=<database> password=<pass>"
	ctx := context.Background()
	client, err := ent.Open("postgres", "postgresql://fly-user:TlIHN_0Ey2N1OKd0a_X-XPjU4o0Gyivj@localhost:16380/kcal-tracker-stag?sslmode=disable")
	if err != nil {
		logger.Panic("opening ent client", zap.Error(err))
	}
	defer client.Close()
	if err := client.Schema.Create(ctx); err != nil {
		logger.Panic("running schema migration", zap.Error(err))
	}

	store := cache.NewInMemoryCache()
	jwkClient := NewJWKSClient(config.ClerkSecret)
	verifier := authClerk.NewVerifier(store, jwkClient, logger)

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

	r.Get("/products/search", func(w http.ResponseWriter, r *http.Request) {
		term := r.URL.Query().Get("q")
		if term == "" {
			http.Error(w, "missing q", http.StatusBadRequest)
			return
		}

		products, err := client.Product.
			Query().
			Where(product.Or(
				product.NameContainsFold(term),
				product.BrandContainsFold(term),
			)).Limit(25).All(r.Context())

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(products)
	})

	log.Printf("listening on :%v", config.Port)
	if err := http.ListenAndServe(":"+strconv.Itoa(config.Port), r); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
