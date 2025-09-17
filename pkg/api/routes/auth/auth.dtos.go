package auth

import "easyflow-oauth2-server/pkg/database"

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	// FirstName and LastName are optional fields
	FirstName *string `json:"firstName,omitempty"`
	LastName  *string `json:"lastName,omitempty"`
}

type CreateUserResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FirstName *string `json:"firstName,omitempty"`
	LastName  *string `json:"lastName,omitempty"`
}

type LoginRequest struct {
	Email        string              `json:"email" binding:"required,email"`
	Password     string              `json:"password" binding:"required"`
	ClientID     string              `json:"clientId" binding:"required"`
	ResponseType database.GrantTypes `json:"grantType" binding:"required,oneof=refresh_token code device_code client_credentials pkce"`
}

type LoginResponse struct {
	SessionToken string `json:"sessionToken"`
	ExpiresIn    int    `json:"expiresIn"`
}
