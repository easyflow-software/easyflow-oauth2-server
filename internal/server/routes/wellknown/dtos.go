package wellknown

// OAuth2Metadata represents the OAuth 2.0 Authorization Server Metadata
// as defined in RFC 8414.
type OAuth2Metadata struct {
	Issuer                                     string   `json:"issuer"`
	AuthorizationEndpoint                      string   `json:"authorization_endpoint"`
	TokenEndpoint                              string   `json:"token_endpoint"`
	JwksURI                                    string   `json:"jwks_uri"`
	ResponseTypesSupported                     []string `json:"response_types_supported"`
	GrantTypesSupported                        []string `json:"grant_types_supported"`
	SubjectTypesSupported                      []string `json:"subject_types_supported,omitempty"`
	ScopesSupported                            []string `json:"scopes_supported,omitempty"`
	TokenEndpointAuthMethodsSupported          []string `json:"token_endpoint_auth_methods_supported"`
	CodeChallengeMethodsSupported              []string `json:"code_challenge_methods_supported,omitempty"`
	ServiceDocumentation                       string   `json:"service_documentation,omitempty"`
	UILocalesSupported                         []string `json:"ui_locales_supported,omitempty"`
	OpPolicyURI                                string   `json:"op_policy_uri,omitempty"`
	OpTosURI                                   string   `json:"op_tos_uri,omitempty"`
	RevocationEndpoint                         string   `json:"revocation_endpoint,omitempty"`
	RevocationEndpointAuthMethodsSupported     []string `json:"revocation_endpoint_auth_methods_supported,omitempty"`
	IntrospectionEndpoint                      string   `json:"introspection_endpoint,omitempty"`
	IntrospectionEndpointAuthMethodsSupported  []string `json:"introspection_endpoint_auth_methods_supported,omitempty"`
	ResponseModesSupported                     []string `json:"response_modes_supported,omitempty"`
	RegistrationEndpoint                       string   `json:"registration_endpoint,omitempty"`
	TokenEndpointAuthSigningAlgValuesSupported []string `json:"token_endpoint_auth_signing_alg_values_supported,omitempty"`
}

// JWKSet represents a JSON Web Key Set as defined in RFC 7517.
type JWKSet struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key as defined in RFC 7517.
type JWK struct {
	KeyType   string `json:"kty"`           // Key Type (e.g., "OKP" for Octet string key pairs)
	Use       string `json:"use,omitempty"` // Public Key Use (e.g., "sig" for signature)
	KeyID     string `json:"kid,omitempty"` // Key ID
	Algorithm string `json:"alg,omitempty"` // Algorithm (e.g., "EdDSA")
	Curve     string `json:"crv,omitempty"` // Curve (e.g., "Ed25519")
	X         string `json:"x,omitempty"`   // X Coordinate (base64url encoded)
	N         string `json:"n,omitempty"`   // Modulus (for RSA)
	E         string `json:"e,omitempty"`   // Exponent (for RSA)
}
