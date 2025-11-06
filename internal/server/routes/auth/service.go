package auth

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"easyflow-oauth2-server/internal/database"
	"easyflow-oauth2-server/internal/errors"
	"easyflow-oauth2-server/internal/helpers"
	"easyflow-oauth2-server/internal/service"
	"easyflow-oauth2-server/internal/tokens"
	e "errors"
	"net/http"
	"time"

	"github.com/lib/pq"
	"go.uber.org/fx"
	"golang.org/x/crypto/bcrypt"
)

// Service handles authentication business logic.
type Service struct {
	*service.BaseService
	Key *ed25519.PrivateKey
}

// ServiceParams holds dependencies for AuthService.
type ServiceParams struct {
	fx.In
	service.BaseServiceParams
	Key *ed25519.PrivateKey
}

// NewAuthService creates a new instance of AuthService.
func NewAuthService(params ServiceParams) *Service {
	baseService := service.NewBaseService("AuthService", params.BaseServiceParams)
	return &Service{
		BaseService: baseService,
		Key:         params.Key,
	}
}

// Register creates a new user account.
func (s *Service) Register(
	ctx context.Context,
	payload CreateUserRequest,
	clientIP string,
) (*CreateUserResponse, *errors.APIError) {
	logger := s.GetLogger(clientIP)
	// Hash the password
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(payload.Password),
		s.Config.SaltRounds,
	)
	if err != nil {
		return nil, &errors.APIError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to hash password",
		}
	}
	logger.PrintfDebug("Successfully hashed password")

	user, err := s.Queries.CreateUser(ctx, database.CreateUserParams{
		Email:        payload.Email,
		PasswordHash: string(hash),
		FirstName:    helpers.StringPtrToNullString(payload.FirstName),
		LastName:     helpers.StringPtrToNullString(payload.LastName),
	})
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			logger.PrintfWarning(
				"Attempted to create user with existing email: %s",
				payload.Email,
			)
			return nil, &errors.APIError{
				Code:    http.StatusConflict,
				Error:   errors.AlreadyExists,
				Details: "Email already in use",
			}
		}
		return nil, &errors.APIError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to create user",
		}
	}
	logger.PrintfInfo("User with id %s created", user.ID)

	return &CreateUserResponse{
		ID:    user.ID.String(),
		Email: user.Email,
		FirstName: func() *string {
			if user.FirstName.Valid {
				return &user.FirstName.String
			}
			return nil
		}(),
		LastName: func() *string {
			if user.LastName.Valid {
				return &user.LastName.String
			}
			return nil
		}(),
	}, nil
}

// Login authenticates a user and returns a session token.
func (s *Service) Login(
	ctx context.Context,
	payload LoginRequest,
	clientIP string,
) (*LoginResponse, *errors.APIError) {
	logger := s.GetLogger(clientIP)
	user, err := s.Queries.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		if e.Is(err, sql.ErrNoRows) {
			logger.PrintfWarning(
				"Attempted login with nonexistent user: %s",
				payload.Email,
			)
			return nil, &errors.APIError{
				Code:    http.StatusUnauthorized,
				Error:   errors.Unauthorized,
				Details: "Invalid email or password",
			}
		}
		logger.PrintfError("Failed to get user by email: %v", err)
		return nil, &errors.APIError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to get user by email",
		}
	}
	logger.PrintfDebug("Found user with email: %s", payload.Email)

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(payload.Password))
	if err != nil {
		logger.PrintfWarning("Invalid password for user: %s", payload.Email)
		return nil, &errors.APIError{
			Code:    http.StatusUnauthorized,
			Error:   errors.Unauthorized,
			Details: "Invalid email or password",
		}
	}
	logger.PrintfDebug("Password for user %s is valid", payload.Email)

	// TODO: Implement more robust client handling with session revocation, etc.
	// For now, we just issue a short-lived session token.
	sessionToken, err := tokens.GenerateSessionToken(s.Config, s.Key, user.ID.String())
	if err != nil {
		logger.PrintfError("Failed to generate session token: %v", err)
		return nil, &errors.APIError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to generate session token",
		}
	}
	logger.PrintfDebug("Generated session token for user %s", payload.Email)

	return &LoginResponse{
		SessionToken: sessionToken,
		ExpiresIn:    s.Config.JwtSessionTokenExpiryHours * int(time.Hour.Seconds()),
	}, nil
}
