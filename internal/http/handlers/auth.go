package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"phsio_track_backend/internal/core"
	"phsio_track_backend/internal/repo"
)

type AuthHandler struct {
	userRepo  *repo.UserRepo
	jwtSecret string
	issuer    string
	expiry    time.Duration
}

func NewAuthHandler(userRepo *repo.UserRepo, jwtSecret, issuer string, expiry time.Duration) *AuthHandler {
	return &AuthHandler{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		issuer:    issuer,
		expiry:    expiry,
	}
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type loginResponse struct {
	Token string `json:"token"`
}

// Login validates credentials and issues a JWT.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	user, err := h.userRepo.GetByUsername(c, req.Username)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"sub": req.Username,
		"iat": now.Unix(),
		"exp": now.Add(h.expiry).Unix(),
		"iss": h.issuer,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token issue failed"})
		return
	}

	c.JSON(http.StatusOK, loginResponse{Token: signed})
}

// SeedUser is a helper to create the first user if needed.
func SeedUser(userRepo *repo.UserRepo, username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user := core.User{
		Username:     username,
		PasswordHash: string(hash),
	}
	return userRepo.UpsertUser(context.Background(), user)
}
