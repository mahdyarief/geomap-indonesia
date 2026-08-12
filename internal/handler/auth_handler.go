package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mahdyarief/geomap-indonesia/internal/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type authRequest struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

// POST /auth
func (h *AuthHandler) Authenticate(c *gin.Context) {
	var req authRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		failure(c, http.StatusBadRequest, codeBadRequest, "invalid request body")
		return
	}
	token, err := h.svc.GenerateToken(req.PublicKey, req.PrivateKey)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			failure(c, http.StatusUnauthorized, codeUnauthorized, "invalid credentials")
			return
		}
		failure(c, http.StatusInternalServerError, codeInternalError, "failed to generate token")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"token":      token.Token,
		"token_type": token.TokenType,
		"expires_in": token.ExpiresIn,
	})
}
