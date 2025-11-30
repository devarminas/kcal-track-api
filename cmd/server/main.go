package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"devarminas/kcal-track-api/ent"
	"devarminas/kcal-track-api/ent/product"
	authClerk "devarminas/kcal-track-api/internal/auth/clerk"
	"devarminas/kcal-track-api/internal/cache"
	"devarminas/kcal-track-api/internal/middleware"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	"github.com/gin-gonic/gin"
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

	router := gin.New()
	router.Use(gin.Recovery())

	if err := router.SetTrustedProxies(nil); err != nil {
		logger.Warn("failed to set trusted proxies", zap.Error(err))
	}

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "welcome")
	})

	router.GET("/test", middleware.ClerkAuth(verifier, logger), func(c *gin.Context) {
		c.String(http.StatusOK, "welcome")
	})

	router.GET("/products/search", func(c *gin.Context) {
		term := c.Query("q")
		if term == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing q"})
			return
		}

		products, err := client.Product.
			Query().
			Where(product.Or(
				product.NameContainsFold(term),
				product.BrandContainsFold(term),
			)).Limit(25).All(c.Request.Context())

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, products)
	})

	log.Printf("listening on :%v", config.Port)
	if err := http.ListenAndServe(":"+strconv.Itoa(config.Port), router); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
