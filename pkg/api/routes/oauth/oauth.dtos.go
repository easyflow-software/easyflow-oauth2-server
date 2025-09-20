package oauth

type AuthorizationCodeTokenResponse struct {
	AccessToken           string   `json:"access_token"`
	AccessTokenExpiresIn  int      `json:"expires_in"` // lifetime in seconds of the access token
	RefreshToken          string   `json:"refresh_token,omitempty"`
	RefreshTokenExpiresIn int      `json:"refresh_token_expires_in,omitempty"` // lifetime in seconds of the refresh token
	Scopes                []string `json:"scopes,omitempty"`
}
