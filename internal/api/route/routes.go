package route

import (
	"gold_portal/config"
	"gold_portal/internal/api/handlers"
	"gold_portal/internal/api/middleware"
	"gold_portal/internal/domain/repositories"
	"gold_portal/internal/domain/services"
	"gold_portal/internal/infrastructure/cache"
	"gold_portal/internal/pkg/jwt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

func SetupRoutes(db *gorm.DB, cfg *config.Config) *gin.Engine {
	router := gin.Default()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "gold-portal",
			"version": "1.0.0",
		})
	})

	router.GET("", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"message": "Gold Portal API",
			"version": "1.0.0",
			"docs":    "/swagger/index.html",
		})
	})

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60, // 12 hours
	}))

	// Repositories
	userRepository := repositories.NewUserRepository(db)

	// Cache (Redis)
	redisCache, err := cache.NewRedisCache(cfg)
	if err != nil {
		panic("Failed to initialize cache: " + err.Error())
	}

	// JWT Service
	jwtService := jwt.NewJWTService(cfg.JWT.Secret)

	// Token Service
	tokenService := services.NewTokenService(redisCache, jwtService)

	// File Service
	fileService, err := services.NewFileService(cfg)
	if err != nil {
		panic("Failed to initialize file service: " + err.Error())
	}

	// Services
	auditService := services.NewAuditService(db)
	authService := services.NewAuthService(userRepository, tokenService, fileService, cfg)
	userService := services.NewUserService(userRepository, fileService)

	// Initialize middleware
	authMiddleware := middleware.AuthMiddleware(authService)
	auditMiddleware := middleware.AuditMiddleware(auditService)
	tokenBlacklistMiddleware := middleware.TokenBlacklistMiddleware(tokenService)
	adminMiddleware := middleware.RequireRoleLevelMiddleware(2)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	auditHandler := handlers.NewAuditHandler(auditService)

	//API routes
	api := router.Group("/api/v1")
	{
		// Public authentication routes (with rate limiting)
		auth := api.Group("/auth")
		auth.Use(auditMiddleware)
		{
			auth.POST("/web-register", authHandler.UserRegister)
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/refresh", authHandler.Refresh)
		}
		authAuth := auth.Group("/")
		authAuth.Use(authMiddleware, auditMiddleware, tokenBlacklistMiddleware)
		{
			authAuth.GET("/me", authHandler.UserMe)
		}

		protected := api.Group("/")
		protected.Use(authMiddleware, tokenBlacklistMiddleware)
		{
			dashboard := protected.Group("/dashboard")
			dashboard.Use(adminMiddleware, auditMiddleware)
			{
				dashboard.GET("", userHandler.GetAll)
				dashboard.POST("register", authHandler.Register)
				dashboard.GET("/id/:id", userHandler.GetByID)
				dashboard.GET("/phone/:phone", userHandler.GetByPhone)
				dashboard.PATCH("/patch/:id", userHandler.Patch)
				dashboard.DELETE("/delete/:id", userHandler.Delete)

			}
		}

	}
	audit := api.Group("/audit")
	audit.Use(authMiddleware, adminMiddleware, tokenBlacklistMiddleware) // только авторизованные админы
	{
		audit.GET("", auditHandler.GetAllLogs)
	}
	return router
}
