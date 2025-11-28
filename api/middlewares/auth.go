package middlewares

import (
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/itsHenry35/SambaManager/config"
	"github.com/itsHenry35/SambaManager/utils"
)

// Claims represents JWT token claims
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// userExistenceCache caches the existence status of users
type userExistenceCache struct {
	mu    sync.RWMutex
	cache map[string]cacheEntry
}

type cacheEntry struct {
	exists    bool
	expiresAt time.Time
}

var userCache = &userExistenceCache{
	cache: make(map[string]cacheEntry),
}

const cacheTTL = 1 * time.Minute

// checkUserExists verifies if a user exists in Samba, with caching
func (uc *userExistenceCache) checkUserExists(username string) bool {
	// Check cache first
	uc.mu.RLock()
	if entry, found := uc.cache[username]; found && time.Now().Before(entry.expiresAt) {
		uc.mu.RUnlock()
		return entry.exists
	}
	uc.mu.RUnlock()

	// Cache miss or expired, check actual user existence
	exists := false

	// Check if it's the admin user
	if username == config.AppConfig.Admin.Username {
		exists = true
	} else {
		// Check regular user via pdbedit
		cmd := exec.Command("pdbedit", "-L", "-u", username)
		if _, err := cmd.CombinedOutput(); err == nil {
			exists = true
		}
	}

	// Update cache
	uc.mu.Lock()
	uc.cache[username] = cacheEntry{
		exists:    exists,
		expiresAt: time.Now().Add(cacheTTL),
	}
	uc.mu.Unlock()

	return exists
}

// AuthMiddleware validates JWT token and sets username in context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ResponseUnauthorized(c, "Authorization header required")
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.ResponseUnauthorized(c, "Invalid authorization format")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return config.GetJWTSecret(), nil
		})

		if err != nil || !token.Valid {
			utils.ResponseUnauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Extract claims
		if claims, ok := token.Claims.(*Claims); ok {
			// Verify user still exists
			if !userCache.checkUserExists(claims.Username) {
				utils.ResponseUnauthorized(c, "User no longer exists")
				c.Abort()
				return
			}

			c.Set("username", claims.Username)
			c.Set("role", claims.Role)
		}

		c.Next()
	}
}

// GetUsernameFromContext retrieves username from context
func GetUsernameFromContext(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}
	usernameStr, ok := username.(string)
	return usernameStr, ok
}

// GetRoleFromContext retrieves role from context
func GetRoleFromContext(c *gin.Context) (string, bool) {
	role, exists := c.Get("role")
	if !exists {
		return "", false
	}
	roleStr, ok := role.(string)
	return roleStr, ok
}

// RequireAdmin middleware ensures only admin can access
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := GetRoleFromContext(c)
		if !exists || role != "admin" {
			utils.ResponseForbidden(c, "Admin access required")
			c.Abort()
			return
		}
		c.Next()
	}
}
