package clerk

import (
	"context"
	"devarminas/kcal-track-api/internal/auth"
	"devarminas/kcal-track-api/internal/cache"
	"errors"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwks"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"go.uber.org/zap"
)

type (
	jwtDecoderFunc  func(ctx context.Context, params *jwt.DecodeParams) (*clerk.UnverifiedToken, error)
	jwkFetcherFunc  func(ctx context.Context, params *jwt.GetJSONWebKeyParams) (*clerk.JSONWebKey, error)
	jwtVerifierFunc func(ctx context.Context, params *jwt.VerifyParams) (*clerk.SessionClaims, error)
)

type ClerkVerifier struct {
	cache     cache.CacheStore
	jwtClient *jwks.Client
	decode    jwtDecoderFunc
	fetchJWK  jwkFetcherFunc
	verifyJWT jwtVerifierFunc
	logger    *zap.Logger
}

func NewVerifier(store cache.CacheStore, client *jwks.Client, logger *zap.Logger) auth.Verifier {
	return newClerkVerifier(store, client, jwt.Decode, jwt.GetJSONWebKey, jwt.Verify, logger)
}

func newClerkVerifier(
	cache cache.CacheStore,
	client *jwks.Client,
	decode jwtDecoderFunc,
	fetchJWK jwkFetcherFunc,
	verifyJWT jwtVerifierFunc,
	logger *zap.Logger,
) *ClerkVerifier {
	return &ClerkVerifier{
		cache:     cache,
		jwtClient: client,
		decode:    decode,
		fetchJWK:  fetchJWK,
		verifyJWT: verifyJWT,
		logger:    logger,
	}
}

func (v *ClerkVerifier) VerifySession(ctx context.Context, token string) (auth.UserClaims, error) {
	jwk, err := v.getOrFetchJWK(ctx, token)
	if err != nil {
		v.logger.Error("failed to fetch JWK during verification", zap.Error(err))
		return auth.UserClaims{}, err
	}

	claims, err := v.verifyJWT(ctx, &jwt.VerifyParams{Token: token, JWK: jwk})
	if err != nil {
		v.logger.Warn("token verification failed", zap.Error(err))
		return auth.UserClaims{}, err
	}

	return auth.UserClaims{Subject: claims.Subject}, nil
}

func (v *ClerkVerifier) getOrFetchJWK(ctx context.Context, token string) (*clerk.JSONWebKey, error) {
	if jwk, err := v.getCachedJWK(); err != nil || jwk != nil {
		return jwk, err
	}

	v.logger.Debug("JWK cache miss, will fetch key")
	return v.fetchAndCacheJWK(ctx, token)
}

const cacheKey = "jwk-key"

func (v *ClerkVerifier) fetchAndCacheJWK(ctx context.Context, token string) (*clerk.JSONWebKey, error) {
	unsafe, err := v.decode(ctx, &jwt.DecodeParams{Token: token})
	if err != nil {
		v.logger.Warn("failed to decode token header", zap.Error(err))
		return nil, err
	}

	jwk, err := v.fetchJWK(ctx, &jwt.GetJSONWebKeyParams{KeyID: unsafe.KeyID, JWKSClient: v.jwtClient})
	if err != nil {
		v.logger.Error("failed to fetch JWK from Clerk", zap.String("kid", unsafe.KeyID), zap.Error(err))
		return nil, err
	}

	v.cache.Set(cacheKey, jwk)
	return jwk, nil
}

func (v *ClerkVerifier) getCachedJWK() (*clerk.JSONWebKey, error) {
	if val, found := v.cache.Get(cacheKey); found {
		if token, ok := val.(*clerk.JSONWebKey); ok {
			return token, nil
		} else {
			return nil, errors.New("failed to cast JSONWebKey")
		}
	}

	return nil, nil
}
