package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/itsHenry35/SambaManager/api/handlers"
	"github.com/itsHenry35/SambaManager/api/middlewares"
)

// SetupRoutes configures all application routes
func SetupRoutes(
	router *gin.Engine,
	userHandler *handlers.UserHandler,
	shareHandler *handlers.ShareHandler,
	userShareHandler *handlers.UserShareHandler,
	userProfileHandler *handlers.UserProfileHandler,
	systemHandler *handlers.SystemHandler,
) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Public routes (no authentication required)
	router.POST("/api/login", handlers.Login)

	// Authenticated routes
	api := router.Group("/api")
	api.Use(middlewares.AuthMiddleware())
	{
		// Admin-only routes
		admin := api.Group("/admin")
		admin.Use(middlewares.RequireAdmin())
		{
			// User management (admin only)
			users := admin.Group("/users")
			{
				users.GET("", userHandler.ListUsers)
				users.GET("/search", userHandler.SearchUsers) // For autocomplete
				users.POST("", userHandler.CreateUser)
				users.DELETE("/:username", userHandler.DeleteUser)
				users.PUT("/:username/password", userHandler.ChangePassword)

				// Orphaned directories management
				users.GET("/orphaned", userHandler.ListOrphanedDirectories)
				users.DELETE("/orphaned/:dirName", userHandler.DeleteOrphanedDirectory)
			}

			// Share management (admin only - full control)
			shares := admin.Group("/shares")
			{
				shares.GET("", shareHandler.ListShares)
				shares.POST("", shareHandler.CreateShare)
				shares.PUT("/:shareId", shareHandler.UpdateShare)
				shares.DELETE("/:shareId", shareHandler.DeleteShare)
			}

			// System management (admin only)
			system := admin.Group("/system")
			{
				system.GET("/check", systemHandler.CheckEnvironment)
				system.GET("/config", systemHandler.GetSambaConfig)
				system.PUT("/config", systemHandler.UpdateSambaConfig)
				system.GET("/config/file", systemHandler.GetSambaConfigFile)
				system.PUT("/config/file", systemHandler.UpdateSambaConfigFile)
				system.GET("/status", systemHandler.GetSambaStatus)
			}
		}

		// User routes (accessible by any authenticated user)
		// All use the same queue as admin routes to prevent concurrent smb.conf access
		user := api.Group("/user")
		{
			// User's own shares
			user.GET("/shares", userShareHandler.ListMyShares)
			user.POST("/shares", userShareHandler.CreateMyShare)
			user.PUT("/shares/:shareId", userShareHandler.UpdateMyShare)
			user.DELETE("/shares/:shareId", userShareHandler.DeleteMyShare)

			// User profile management
			user.PUT("/password", userProfileHandler.ChangeOwnPassword)

			// User search (for sharing purposes)
			user.GET("/users/search", userHandler.SearchUsers)
		}
	}
}
