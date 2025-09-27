package auth

type createUserRequest struct {
	Email    string `json:"email"               binding:"required,email"`
	Password string `json:"password"            binding:"required,min=8"`
	// FirstName and LastName are optional fields
	FirstName *string `json:"firstName,omitempty"`
	LastName  *string `json:"lastName,omitempty"`
}

type createUserResponse struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FirstName *string `json:"firstName,omitempty"`
	LastName  *string `json:"lastName,omitempty"`
}

type loginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type loginResponse struct {
	SessionToken string `json:"sessionToken"`
	ExpiresIn    int    `json:"expiresIn"`
}
