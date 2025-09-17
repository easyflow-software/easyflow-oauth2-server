package auth

import (
	"database/sql"
	"easyflow-oauth2-server/pkg/api/errors"
	"easyflow-oauth2-server/pkg/database"
	"easyflow-oauth2-server/pkg/endpoint"
	"easyflow-oauth2-server/pkg/tokens"
	"net/http"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func register(utils endpoint.EndpointUtils[CreateUserRequest]) (*CreateUserResponse, *errors.ApiError) {
	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(utils.Payload.Password), utils.Config.SaltRounds)
	if err != nil {
		return nil, &errors.ApiError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to hash password",
		}
	}
	utils.Logger.PrintfDebug("Sucessfully hashed password")

	user, err := utils.Queries.CreateUser(utils.RequestContext, database.CreateUserParams{
		Email:        utils.Payload.Email,
		PasswordHash: string(hash),
		FirstName:    sql.NullString{String: *utils.Payload.FirstName, Valid: utils.Payload.FirstName != nil},
		LastName:     sql.NullString{String: *utils.Payload.LastName, Valid: utils.Payload.LastName != nil},
	})
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == "unique_violation" {
			utils.Logger.PrintfWarning("Attempted to create user with existing email: %s", utils.Payload.Email)
			return nil, &errors.ApiError{
				Code:    http.StatusConflict,
				Error:   errors.AlreadyExists,
				Details: "Email already in use",
			}
		}
		return nil, &errors.ApiError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to create user",
		}
	}
	utils.Logger.PrintfInfo("User with id %s created", user.ID)

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

func login(utils endpoint.EndpointUtils[LoginRequest]) (*LoginResponse, *errors.ApiError) {
	user, err := utils.Queries.GetUserByEmail(utils.RequestContext, utils.Payload.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.Logger.PrintfWarning("Attempted login with nonexistent user: %s", utils.Payload.Email)
			return nil, &errors.ApiError{
				Code:    http.StatusUnauthorized,
				Error:   errors.Unauthorized,
				Details: "Invalid email or password",
			}
		}
		utils.Logger.PrintfError("Failed to get user by email: %v", err)
		return nil, &errors.ApiError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to get user by email",
		}
	}
	utils.Logger.PrintfDebug("Found user with email: %s", utils.Payload.Email)

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(utils.Payload.Password))
	if err != nil {
		utils.Logger.PrintfWarning("Invalid password for user: %s", utils.Payload.Email)
		return nil, &errors.ApiError{
			Code:    http.StatusUnauthorized,
			Error:   errors.Unauthorized,
			Details: "Invalid email or password",
		}
	}
	utils.Logger.PrintfDebug("Password for user %s is valid", utils.Payload.Email)

	// TODO: Implement more robust client handling with session revocation, etc.
	// For now, we just issue a short-lived session token.
	sessionToken, err := tokens.GenerateSessionToken(utils.Config, utils.Key, user)

	return &LoginResponse{
		SessionToken: sessionToken,
		ExpiresIn:    utils.Config.JwtSessionTokenExpiryHours * int(time.Hour.Seconds()),
	}, nil
}
