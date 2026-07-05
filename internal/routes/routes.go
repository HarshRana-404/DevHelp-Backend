// Package routes wires all handler registrations onto the Gin engine.
// Adding a new tool only requires registering its handler group here.
package routes

import (
	"time"

	"devhelp/internal/config"
	"devhelp/internal/handlers"
	"devhelp/internal/middleware"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// Setup registers all routes and middleware on the provided Gin engine.
func Setup(r *gin.Engine, cfg *config.Config, log *zap.Logger,
	healthH *handlers.HealthHandler,
	httpInspectorH *handlers.HTTPInspectorHandler,
	curlConverterH *handlers.CurlConverterHandler,
	jwtH *handlers.JWTHandler,
	jsonGenH *handlers.JSONGeneratorHandler,
) {
	// Global middleware — applied to every request.
	r.Use(middleware.Recovery(log))
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(log))
	r.Use(middleware.CORS(cfg.CORS.AllowedOrigins))
	r.Use(middleware.RateLimiter(cfg.RateLimiter.RequestsPerMinute))
	r.Use(middleware.Timeout(time.Duration(cfg.App.TimeoutSeconds) * time.Second))
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	v1 := r.Group("/api/v1")

	// Health
	v1.GET("/health", healthH.Check)

	// HTTP Inspector
	v1.POST("/http-inspector", httpInspectorH.Inspect)

	// cURL Converter
	curl := v1.Group("/curl-converter")
	{
		curl.POST("", curlConverterH.Convert)
		curl.GET("/languages", curlConverterH.Languages)
	}

	// JWT Toolkit
	jwt := v1.Group("/jwt")
	{
		jwt.POST("/generate", jwtH.Generate)
		jwt.POST("/decode", jwtH.Decode)
	}

	// JSON Generator
	v1.POST("/json-generator", jsonGenH.Generate)

	// Swagger UI — available in all environments
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
