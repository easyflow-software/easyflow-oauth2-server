package auth

// CreateUserRequest represents the payload for creating a new user.
type CreateUserRequest struct {
	Email    string `json:"email"               binding:"required,email"`
	Password string `json:"password"            binding:"required,min=8"`
	// FirstName and LastName are optional fields
	FirstName *string `json:"firstName,omitempty"`
	LastName  *string `json:"lastName,omitempty"`
}

// CreateUserResponse represents the response after creating a new user.
type CreateUserResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FirstName *string `json:"firstName,omitempty"`
	LastName  *string `json:"lastName,omitempty"`
}

// LoginRequest represents the payload for user login.
type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the response after a successful login.
type LoginResponse struct {
	SessionToken string `json:"sessionToken"`
	ExpiresIn    int    `json:"expiresIn"`
}
