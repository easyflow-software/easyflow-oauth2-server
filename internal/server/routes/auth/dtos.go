package auth

// CreateUserRequest represents the payload for creating a new user.
type CreateUserRequest struct {
	Email    string `json:"email"                validate:"required,email"`
	Password string `json:"password"             validate:"required,min=8"`
	// FirstName and LastName are optional fields
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
}

// CreateUserResponse represents the response after creating a new user.
type CreateUserResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
}

// LoginRequest represents the payload for user login.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the response after a successful login.
type LoginResponse struct {
	SessionToken string `json:"session_token"`
	ExpiresIn    int    `json:"expiresIn"`
}
