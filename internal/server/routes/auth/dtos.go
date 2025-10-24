package auth

// CreateUserRequest represents the payload for creating a new user.
type CreateUserRequest struct {
	Email    string `json:"email"                validate:"required,email" example:"user@example.com"`  // User's email address
	Password string `json:"password"             validate:"required,min=8" example:"securePassword123"` // User's password (minimum 8 characters)
	// FirstName and LastName are optional fields
	FirstName *string `json:"first_name,omitempty"                           example:"John"` // User's first name (optional)
	LastName  *string `json:"last_name,omitempty"                            example:"Doe"`  // User's last name (optional)
}

// CreateUserResponse represents the response after creating a new user.
type CreateUserResponse struct {
	ID        string  `json:"id"                   example:"550e8400-e29b-41d4-a716-446655440000"` // User's unique identifier
	Email     string  `json:"email"                example:"user@example.com"`                     // User's email address
	FirstName *string `json:"first_name,omitempty" example:"John"`                                 // User's first name
	LastName  *string `json:"last_name,omitempty"  example:"Doe"`                                  // User's last name
}

// LoginRequest represents the payload for user login.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email" example:"user@example.com"`  // User's email address
	Password string `json:"password" validate:"required"       example:"securePassword123"` // User's password
}

// LoginResponse represents the response after a successful login.
type LoginResponse struct {
	SessionToken string `json:"session_token" example:"eyJhbGciOiJFZERTQSIsInR5cCI6IkpXVCJ9..."` // JWT session token
	ExpiresIn    int    `json:"expiresIn"     example:"3600"`                                    // Token expiration time in seconds
}
