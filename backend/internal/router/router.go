package router

import (
	"github.com/gin-gonic/gin"
	"github.com/hdu-dp/backend/internal/handlers"
	adminHandlers "github.com/hdu-dp/backend/internal/handlers/admin"
	"github.com/hdu-dp/backend/internal/middleware"
)

// Params groups dependencies required for routing.
type Params struct {
	Engine          *gin.Engine
	AuthMiddleware  *middleware.AuthMiddleware
	AuthHandler     *handlers.AuthHandler
	UserHandler     *handlers.UserHandler
	ReviewHandler   *handlers.ReviewHandler
	AdminHandler    *adminHandlers.ReviewAdminHandler
	StaticUploadDir string
}

// Register configures API routes on the provided engine.
func Register(p Params) {
	api := p.Engine.Group("/api/v1")

	auth := api.Group("/auth")
	{
		auth.POST("/register", p.AuthHandler.Register)
		auth.POST("/login", p.AuthHandler.Login)
		auth.POST("/refresh", p.AuthHandler.Refresh)
		auth.POST("/logout", p.AuthHandler.Logout)
	}

	if p.StaticUploadDir != "" {
		api.Static("/uploads", p.StaticUploadDir)
	}

	api.GET("/reviews", p.ReviewHandler.ListPublic)
	// Detail endpoint should be accessible to authed/unauthed; optional auth ensures role-based access when provided.
	api.GET("/reviews/:id", p.AuthMiddleware.OptionalAuth(), p.ReviewHandler.Detail)

	protected := api.Group("")
	protected.Use(p.AuthMiddleware.RequireAuth())
	{
		protected.GET("/users/me", p.UserHandler.Me)

		protected.POST("/reviews", p.ReviewHandler.Submit)
		protected.GET("/reviews/me", p.ReviewHandler.MyReviews)
		protected.POST("/reviews/:id/images", p.ReviewHandler.UploadImage)
	}

	admin := api.Group("/admin")
	admin.Use(p.AuthMiddleware.RequireAuth(), p.AuthMiddleware.RequireRoles("admin"))
	{
		admin.GET("/reviews/pending", p.AdminHandler.Pending)
		admin.PUT("/reviews/:id/approve", p.AdminHandler.Approve)
		admin.PUT("/reviews/:id/reject", p.AdminHandler.Reject)
		admin.DELETE("/reviews/:id", p.AdminHandler.Delete)
	}
}
