package oauth

// TokenResponse represents the response returned after a successful token request.
type TokenResponse struct {
	AccessToken           string   `json:"access_token"`
	AccessTokenExpiresIn  int      `json:"expires_in"` // lifetime in seconds of the access token
	RefreshToken          string   `json:"refresh_token,omitempty"`
	RefreshTokenExpiresIn int      `json:"refresh_token_expires_in,omitempty"` // lifetime in seconds of the refresh token
	Scopes                []string `json:"scopes"`
}
