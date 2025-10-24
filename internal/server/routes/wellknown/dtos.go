package wellknown

// OAuth2Metadata represents the OAuth 2.0 Authorization Server Metadata
// as defined in RFC 8414.
type OAuth2Metadata struct {
	Issuer                                     string   `json:"issuer"                                                     example:"https://auth.easyflow.com"`                       // OAuth2 issuer identifier
	AuthorizationEndpoint                      string   `json:"authorization_endpoint"                                     example:"https://auth.easyflow.com/oauth/authorize"`       // Authorization endpoint URL
	TokenEndpoint                              string   `json:"token_endpoint"                                             example:"https://auth.easyflow.com/oauth/token"`           // Token endpoint URL
	JwksURI                                    string   `json:"jwks_uri"                                                   example:"https://auth.easyflow.com/.well-known/jwks.json"` // JSON Web Key Set URI
	ResponseTypesSupported                     []string `json:"response_types_supported"                                   example:"code"`                                            // Supported OAuth2 response types
	GrantTypesSupported                        []string `json:"grant_types_supported"                                      example:"authorization_code,refresh_token"`                // Supported OAuth2 grant types
	SubjectTypesSupported                      []string `json:"subject_types_supported,omitempty"                          example:"public"`                                          // Supported subject types
	ScopesSupported                            []string `json:"scopes_supported,omitempty"                                 example:"openid,profile,email"`                            // Supported scopes
	TokenEndpointAuthMethodsSupported          []string `json:"token_endpoint_auth_methods_supported"                      example:"client_secret_basic,client_secret_post"`          // Supported token endpoint authentication methods
	CodeChallengeMethodsSupported              []string `json:"code_challenge_methods_supported,omitempty"                 example:"S256"`                                            // Supported PKCE code challenge methods
	ServiceDocumentation                       string   `json:"service_documentation,omitempty"                            example:"https://docs.easyflow.com"`                       // Service documentation URL
	UILocalesSupported                         []string `json:"ui_locales_supported,omitempty"                             example:"en-US,de-DE"`                                     // Supported UI locales
	OpPolicyURI                                string   `json:"op_policy_uri,omitempty"                                    example:"https://easyflow.com/policy"`                     // Operator policy URI
	OpTosURI                                   string   `json:"op_tos_uri,omitempty"                                       example:"https://easyflow.com/tos"`                        // Operator terms of service URI
	RevocationEndpoint                         string   `json:"revocation_endpoint,omitempty"                              example:"https://auth.easyflow.com/oauth/revoke"`          // Token revocation endpoint
	RevocationEndpointAuthMethodsSupported     []string `json:"revocation_endpoint_auth_methods_supported,omitempty"       example:"client_secret_basic"`                             // Supported revocation endpoint auth methods
	IntrospectionEndpoint                      string   `json:"introspection_endpoint,omitempty"                           example:"https://auth.easyflow.com/oauth/introspect"`      // Token introspection endpoint
	IntrospectionEndpointAuthMethodsSupported  []string `json:"introspection_endpoint_auth_methods_supported,omitempty"    example:"client_secret_basic"`                             // Supported introspection endpoint auth methods
	ResponseModesSupported                     []string `json:"response_modes_supported,omitempty"                         example:"query,fragment"`                                  // Supported response modes
	RegistrationEndpoint                       string   `json:"registration_endpoint,omitempty"                            example:"https://auth.easyflow.com/oauth/register"`        // Dynamic client registration endpoint
	TokenEndpointAuthSigningAlgValuesSupported []string `json:"token_endpoint_auth_signing_alg_values_supported,omitempty" example:"RS256,ES256"`                                     // Supported signing algorithms for token endpoint auth
}

// JWKSet represents a JSON Web Key Set as defined in RFC 7517.
type JWKSet struct {
	Keys []JWK `json:"keys"` // Array of JSON Web Keys
}

// JWK represents a JSON Web Key as defined in RFC 7517.
type JWK struct {
	KeyType   string `json:"kty"           example:"OKP"`                                         // Key Type (e.g., "OKP" for Octet string key pairs)
	Use       string `json:"use,omitempty" example:"sig"`                                         // Public Key Use (e.g., "sig" for signature)
	KeyID     string `json:"kid,omitempty" example:"key-1"`                                       // Key ID
	Algorithm string `json:"alg,omitempty" example:"EdDSA"`                                       // Algorithm (e.g., "EdDSA")
	Curve     string `json:"crv,omitempty" example:"Ed25519"`                                     // Curve (e.g., "Ed25519")
	X         string `json:"x,omitempty"   example:"11qYAYKxCrfVS_7TyWQHOg7hcvPapiMlrwIaaPcHURo"` // X Coordinate (base64url encoded)
	N         string `json:"n,omitempty"`                                                         // Modulus (for RSA)
	E         string `json:"e,omitempty"`                                                         // Exponent (for RSA)
}
