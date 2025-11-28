package handlers

import (
	"os/exec"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/itsHenry35/SambaManager/api/middlewares"
	"github.com/itsHenry35/SambaManager/config"
	"github.com/itsHenry35/SambaManager/types"
	"github.com/itsHenry35/SambaManager/utils"
)

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response data
type LoginResponse struct {
	Token     string         `json:"token"`
	Username  string         `json:"username"`
	Role      types.UserRole `json:"role"`
	ExpiresAt int64          `json:"expires_at"`
}

// Login authenticates a user and returns a JWT token
// Supports both admin (from config) and regular users (via Samba authentication)
func Login(c *gin.Context) {
	var credentials LoginRequest

	if err := c.ShouldBindJSON(&credentials); err != nil {
		utils.ResponseBadRequest(c, err.Error())
		return
	}

	// SECURITY: Validate username format to prevent command injection
	// Only allow alphanumeric, underscore, dash (1-32 chars)
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]{1,32}$`)
	if !usernameRegex.MatchString(credentials.Username) {
		utils.ResponseUnauthorized(c, "Illegal request")
		return
	}

	var role types.UserRole

	// Check if admin login
	if credentials.Username == config.AppConfig.Admin.Username &&
		credentials.Password == config.AppConfig.Admin.Password {
		role = types.RoleAdmin
	} else {
		// Try to authenticate as regular user via Samba
		// First check if user exists in Samba (username already validated above)
		cmd := exec.Command("pdbedit", "-L", "-u", credentials.Username)
		if _, err := cmd.CombinedOutput(); err != nil {
			utils.ResponseUnauthorized(c, "Invalid credentials")
			return
		}

		// Verify password by attempting to authenticate with smbclient
		// SECURITY: username is validated with regex above
		// password is passed as part of -U parameter in the format "username%password"
		// This is safe because smbclient handles special characters in passwords correctly
		cmd = exec.Command("smbclient", "-L", "localhost", "-U", credentials.Username+"%"+credentials.Password)
		if err := cmd.Run(); err != nil {
			utils.ResponseUnauthorized(c, "Invalid credentials")
			return
		}

		role = types.RoleUser
	}

	// Generate JWT token
	expirationTime := time.Now().Add(24 * 30 * time.Hour)
	claims := &middlewares.Claims{
		Username: credentials.Username,
		Role:     string(role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(config.GetJWTSecret())
	if err != nil {
		utils.ResponseInternalServerError(c, "Failed to generate token")
		return
	}

	utils.ResponseOK(c, LoginResponse{
		Token:     tokenString,
		Username:  credentials.Username,
		Role:      role,
		ExpiresAt: expirationTime.Unix(),
	})
}
