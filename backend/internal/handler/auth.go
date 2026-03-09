package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/lobsterpool/lobsterpool/internal/models"
	"golang.org/x/crypto/bcrypt"
	moderncsqlite "modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

type AuthHandler struct {
	userStore *models.UserStore
	jwtSecret []byte
}

const (
	authInternalError           = "internal error"
	authInvalidCredentialsError = "invalid username or password"
	authInvalidNewPasswordError = "new password is required"
	authRequiredFieldsError     = "username and password are required"
	authUnauthorizedError       = "unauthorized"
	authUserNotFoundError       = "user not found"
	authUsernameExistsError     = "username already exists"
	loggedOutMessage            = "logged out"
)

func NewAuthHandler(userStore *models.UserStore, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		userStore: userStore,
		jwtSecret: []byte(jwtSecret),
	}
}

type credentialsRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type authResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

type changePasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required"`
}

func isUniqueConstraintError(err error) bool {
	var sqliteErr *moderncsqlite.Error
	if !errors.As(err, &sqliteErr) {
		return false
	}

	return sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE
}

func (h *AuthHandler) bindCredentials(c *gin.Context) (*credentialsRequest, bool) {
	var req credentialsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, authRequiredFieldsError)
		return nil, false
	}

	return &req, true
}

func (h *AuthHandler) respondWithToken(c *gin.Context, status int, user *models.User) {
	token, err := h.generateToken(user.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, authInternalError)
		return
	}

	c.JSON(status, authResponse{Token: token, User: user})
}

func (h *AuthHandler) Register(c *gin.Context) {
	req, ok := h.bindCredentials(c)
	if !ok {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(c, http.StatusInternalServerError, authInternalError)
		return
	}

	user := &models.User{
		ID:                 uuid.New().String(),
		Username:           req.Username,
		Role:               models.UserRoleMember,
		MustChangePassword: false,
		PasswordHash:       string(hash),
	}

	if err := h.userStore.Create(user); err != nil {
		if isUniqueConstraintError(err) {
			respondError(c, http.StatusConflict, authUsernameExistsError)
			return
		}

		respondError(c, http.StatusInternalServerError, authInternalError)
		return
	}

	h.respondWithToken(c, http.StatusCreated, user)
}

func (h *AuthHandler) Login(c *gin.Context) {
	req, ok := h.bindCredentials(c)
	if !ok {
		return
	}

	user, err := h.userStore.GetByUsername(req.Username)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		respondError(c, http.StatusUnauthorized, authInvalidCredentialsError)
		return
	case err != nil:
		respondError(c, http.StatusInternalServerError, authInternalError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		respondError(c, http.StatusUnauthorized, authInvalidCredentialsError)
		return
	}

	h.respondWithToken(c, http.StatusOK, user)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok || userID == "" {
		respondError(c, http.StatusUnauthorized, authUnauthorizedError)
		return
	}

	user, err := h.userStore.GetByID(userID)
	if err != nil {
		respondError(c, http.StatusUnauthorized, authUserNotFoundError)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	respondMessage(c, http.StatusOK, loggedOutMessage)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok || userID == "" {
		respondError(c, http.StatusUnauthorized, authUnauthorizedError)
		return
	}

	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.NewPassword == "" {
		respondError(c, http.StatusBadRequest, authInvalidNewPasswordError)
		return
	}

	user, err := h.userStore.GetByID(userID)
	if err != nil {
		respondError(c, http.StatusUnauthorized, authUserNotFoundError)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		respondError(c, http.StatusInternalServerError, authInternalError)
		return
	}

	if err := h.userStore.UpdatePassword(user.ID, string(hash), false); err != nil {
		respondError(c, http.StatusInternalServerError, authInternalError)
		return
	}

	updatedUser, err := h.userStore.GetByID(userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, authInternalError)
		return
	}

	c.JSON(http.StatusOK, updatedUser)
}

func (h *AuthHandler) generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(h.jwtSecret)
}
