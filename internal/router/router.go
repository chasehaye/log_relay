package router

import (
	"log"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

	"log_relay/internal/config"
    "log_relay/internal/middleware"
    "log_relay/internal/router/routes"

    _ "log_relay/docs"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

func Setup(db *gorm.DB) *gin.Engine {
    if config.IsProduction() {
        log.Println("Running in Production mode")
        gin.SetMode(gin.ReleaseMode)
    } else {
        log.Println("Running in Development mode")
        gin.SetMode(gin.DebugMode)
    }

    r := gin.New()
    r.SetTrustedProxies([]string{"127.0.0.1", "::1"})
    r.Use(gin.Recovery())
    if !config.IsProduction() {
        r.Use(gin.Logger())
    }

    r.Use(middleware.CORSMiddleware())

    routes.Public(r, db)
    routes.Semi(r, db)
    routes.Protected(r, db)

    if !config.IsProduction() {
        r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    }

    return r
}