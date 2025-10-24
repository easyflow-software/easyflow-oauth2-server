package oauth

// TokenResponse represents the response returned after a successful token request.
type TokenResponse struct {
	AccessToken           string   `json:"access_token"                       example:"eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9..."` // OAuth2 access token
	AccessTokenExpiresIn  int      `json:"expires_in"                         example:"3600"`                                    // Lifetime in seconds of the access token
	RefreshToken          string   `json:"refresh_token,omitempty"            example:"eyJhbGciOiJFZERTQSIsInR5cCI6..."`         // OAuth2 refresh token (optional)
	RefreshTokenExpiresIn int      `json:"refresh_token_expires_in,omitempty" example:"86400"`                                   // Lifetime in seconds of the refresh token (optional)
	Scopes                []string `json:"scopes"                             example:"read,write"`                              // Granted scopes
}
