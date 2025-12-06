package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"devarminas/kcal-track-api/ent"
	v1 "devarminas/kcal-track-api/internal/api/v1"
	authClerk "devarminas/kcal-track-api/internal/auth/clerk"
	"devarminas/kcal-track-api/internal/cache"
	domainproduct "devarminas/kcal-track-api/internal/domain/product"
	"devarminas/kcal-track-api/internal/middleware"

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
	productRepo := domainproduct.NewRepository(client)
	jwkClient := NewJWKSClient(config.ClerkSecret)
	verifier := authClerk.NewVerifier(store, jwkClient, logger)
	productHandler := v1.NewProductHandler(productRepo)

	router := gin.New()
	router.Use(gin.Recovery())

	if err := router.SetTrustedProxies(nil); err != nil {
		logger.Warn("failed to set trusted proxies", zap.Error(err))
	}

	products := router.Group("/products")
	products.Use(middleware.ClerkAuth(verifier, logger))
	products.GET("/search", productHandler.SearchProducts)
	products.GET("/barcode/:barcode", productHandler.GetProductByBarcode)
	products.POST("", productHandler.CreateProduct)

	log.Printf("listening on :%v", config.Port)
	if err := http.ListenAndServe(":"+strconv.Itoa(config.Port), router); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
