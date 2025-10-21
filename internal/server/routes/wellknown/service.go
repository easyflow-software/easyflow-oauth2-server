package wellknown

import (
	"context"
	"crypto/ed25519"
	"easyflow-oauth2-server/internal/shared/service"
	"easyflow-oauth2-server/pkg/database"
	"encoding/base64"
	"fmt"

	"go.uber.org/fx"
)

// Service handles well-known endpoint business logic.
type Service struct {
	*service.BaseService
	key *ed25519.PrivateKey
}

// ServiceParams holds dependencies for WellKnownService.
type ServiceParams struct {
	fx.In
	service.BaseServiceParams
	Key *ed25519.PrivateKey
}

// NewWellKnownService creates a new instance of WellKnownService.
func NewWellKnownService(deps ServiceParams) *Service {
	baseService := service.NewBaseService("WellKnownService", deps.BaseServiceParams)
	return &Service{
		BaseService: baseService,
		key:         deps.Key,
	}
}

// GetOAuth2Metadata generates the OAuth 2.0 Authorization Server Metadata.
func (s *Service) GetOAuth2Metadata(ctx context.Context, clientIP string) *OAuth2Metadata {
	logger := s.GetLogger(clientIP)

	// Build base URL from config
	baseURL := s.Config.BaseURL
	if baseURL == "" {
		logger.PrintfWarning("BASE_URL not configured, using default")
		baseURL = fmt.Sprintf("http://localhost:%s", s.Config.Port)
	}

	// Get all supported scopes from database
	scopes, err := s.Queries.ListScopes(ctx)
	scopeNames := []string{}
	if err != nil {
		logger.PrintfWarning("Failed to retrieve scopes from database: %v", err)
	} else {
		for _, scope := range scopes {
			scopeNames = append(scopeNames, scope.Name)
		}
	}

	metadata := &OAuth2Metadata{
		Issuer:                baseURL,
		AuthorizationEndpoint: fmt.Sprintf("%s/oauth/authorize", baseURL),
		TokenEndpoint:         fmt.Sprintf("%s/oauth/token", baseURL),
		JwksURI:               fmt.Sprintf("%s/.well-known/jwks.json", baseURL),
		ResponseTypesSupported: []string{
			"code",
		},
		GrantTypesSupported: []string{
			string(database.GrantTypesAuthorizationCode),
			string(database.GrantTypesClientCredentials),
			string(database.GrantTypesRefreshToken),
		},
		SubjectTypesSupported: []string{
			"public",
		},
		ScopesSupported: scopeNames,
		TokenEndpointAuthMethodsSupported: []string{
			"client_secret_basic",
			"client_secret_post",
			"none", // for public clients
		},
		CodeChallengeMethodsSupported: []string{
			"S256",
		},
		ResponseModesSupported: []string{
			"query",
			"fragment",
		},
		TokenEndpointAuthSigningAlgValuesSupported: []string{
			"EdDSA",
		},
	}

	return metadata
}

// GetJWKS generates the JSON Web Key Set containing the public key.
func (s *Service) GetJWKS(clientIP string) *JWKSet {
	logger := s.GetLogger(clientIP)

	// Extract the public key from the private key
	publicKey := s.key.Public().(ed25519.PublicKey)

	// Encode the public key in base64url format (without padding)
	x := base64.RawURLEncoding.EncodeToString(publicKey)

	logger.PrintfDebug("Generated JWKS with Ed25519 public key")

	return &JWKSet{
		Keys: []JWK{
			{
				KeyType:   "OKP",     // Octet string key pairs
				Use:       "sig",     // Used for signatures
				Algorithm: "EdDSA",   // EdDSA algorithm
				Curve:     "Ed25519", // Ed25519 curve
				X:         x,         // Public key coordinate
			},
		},
	}
}
