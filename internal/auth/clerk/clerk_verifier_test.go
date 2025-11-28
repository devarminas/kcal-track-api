package clerk

import (
	"context"
	"errors"
	"testing"

	"devarminas/kcal-track-api/internal/cache"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestVerifySessionUsesCachedJWK(t *testing.T) {
	expectedSubject := "user-123"
	cachedJWK := &clerk.JSONWebKey{
		Key:       []byte("secret"),
		KeyID:     "kid-123",
		Algorithm: "alg",
		Use:       "sig",
	}

	store := cache.NewInMemoryCache()
	store.Set(cacheKey, cachedJWK)
	token := "test-token"
	verifier := newClerkVerifier(
		store,
		nil,
		func(ctx context.Context, params *jwt.DecodeParams) (*clerk.UnverifiedToken, error) {
			t.Fatalf("unexpected decode call with %v", params)
			return nil, nil
		},
		func(ctx context.Context, params *jwt.GetJSONWebKeyParams) (*clerk.JSONWebKey, error) {
			t.Fatalf("unexpected jwk fetch with %v", params)
			return nil, nil
		},
		func(ctx context.Context, params *jwt.VerifyParams) (*clerk.SessionClaims, error) {
			assert.Equal(t, token, params.Token, "verify called with wrong token")
			require.NotNil(t, params.JWK, "verify called with nil JWK")
			assert.Equal(t, cachedJWK, params.JWK, "verify called with unexpected JWK")
			return &clerk.SessionClaims{
				RegisteredClaims: clerk.RegisteredClaims{Subject: expectedSubject},
			}, nil
		},
		zap.NewNop(),
	)

	claims, err := verifier.VerifySession(context.Background(), token)
	require.NoError(t, err, "verify session returned error")
	assert.Equal(t, expectedSubject, claims.Subject, "unexpected subject")
}

func TestVerifySessionFetchesAndCachesJWK(t *testing.T) {
	token := "new-token"
	keyID := "kid-456"
	fetchedKey := &clerk.JSONWebKey{
		Key:       []byte("secret"),
		KeyID:     keyID,
		Algorithm: "alg",
		Use:       "sig",
	}

	store := cache.NewInMemoryCache()

	verifier := newClerkVerifier(
		store,
		nil,
		func(ctx context.Context, params *jwt.DecodeParams) (*clerk.UnverifiedToken, error) {
			assert.Equal(t, token, params.Token, "decode called with wrong token")
			return &clerk.UnverifiedToken{KeyID: keyID}, nil
		},
		func(ctx context.Context, params *jwt.GetJSONWebKeyParams) (*clerk.JSONWebKey, error) {
			assert.Equal(t, keyID, params.KeyID, "fetch called with wrong keyID")
			return fetchedKey, nil
		},
		func(ctx context.Context, params *jwt.VerifyParams) (*clerk.SessionClaims, error) {
			assert.Equal(t, fetchedKey, params.JWK, "verify called with unexpected JWK")
			assert.Equal(t, token, params.Token, "verify called with wrong token")
			return &clerk.SessionClaims{
				RegisteredClaims: clerk.RegisteredClaims{Subject: "user-456"},
			}, nil
		},
		zap.NewNop(),
	)

	claims, err := verifier.VerifySession(context.Background(), token)
	require.NoError(t, err, "verify session returned error")

	assert.Equal(t, "user-456", claims.Subject, "unexpected subject")

	cached, ok := store.Get(cacheKey)
	require.True(t, ok, "expected fetched JWK to be cached")

	parsed, ok := cached.(*clerk.JSONWebKey)
	require.True(t, ok, "cached JWK stored with wrong type")
	assert.Equal(t, keyID, parsed.KeyID, "cached JWK key id mismatch")
}

func TestVerifySessionReturnsUnauthorizedOnDecodeError(t *testing.T) {
	store := cache.NewInMemoryCache()
	expectedErr := "decode failed"

	verifier := newClerkVerifier(
		store,
		nil,
		func(ctx context.Context, params *jwt.DecodeParams) (*clerk.UnverifiedToken, error) {
			return nil, errors.New(expectedErr)
		},
		func(ctx context.Context, params *jwt.GetJSONWebKeyParams) (*clerk.JSONWebKey, error) {
			t.Fatalf("unexpected jwk fetch")
			return nil, nil
		},
		func(ctx context.Context, params *jwt.VerifyParams) (*clerk.SessionClaims, error) {
			t.Fatalf("unexpected verify call")
			return nil, nil
		},
		zap.NewNop(),
	)

	_, err := verifier.VerifySession(context.Background(), "bad-token")
	require.Error(t, err, "expected decode error")
	assert.EqualError(t, err, expectedErr, "unexpected error message")
}
