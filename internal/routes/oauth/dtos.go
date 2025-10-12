package oauth

type tokenResponse struct {
	AccessToken           string   `json:"access_token"`
	AccessTokenExpiresIn  int      `json:"expires_in"` // lifetime in seconds of the access token
	RefreshToken          string   `json:"refresh_token,omitempty"`
	RefreshTokenExpiresIn int      `json:"refresh_token_expires_in,omitempty"` // lifetime in seconds of the refresh token
	Scopes                []string `json:"scopes"`
}

type registerRequest struct {
	ClientName                     string    `json:"client_name"                                 validate:"required,min=3,max=100"`
	Description                    string    `json:"description"                                 validate:"max=255"`
	RedirectURIs                   []string  `json:"redirect_uris"                               validate:"required,dive,uri"`
	Scopes                         []string  `json:"scopes"                                      validate:"dive,required,lowercase,alphanum"`
	GrantTypes                     *[]string `json:"grant_types"                                 validate:"omitempty,dive,required,oneof=authorization_code client_credentials refresh_token"`
	IsPublic                       bool      `json:"is_public"                                   validate:"required,boolean"`
	AuthorizationCodeValidDuration *int      `json:"authorization_code_valid_duration,omitempty" validate:"omitempty,min=60,max=1800"`     // in seconds
	AccessTokenValidDuration       *int      `json:"access_token_valid_duration,omitempty"       validate:"omitempty,min=300,max=86400"`   // in seconds
	RefreshTokenValidDuration      *int      `json:"refresh_token_valid_duration,omitempty"      validate:"omitempty,min=3600,max=604800"` // in seconds
}
