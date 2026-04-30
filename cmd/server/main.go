// @title           Log Relay API
// @version         1.0
// @description     API Server for Log Relay Project.
// @host            localhost:8080
// @BasePath        /
package main

import (
    "log"
    "os"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"

    "log_relay/internal/database"
    "log_relay/internal/models"
    "log_relay/internal/middleware"
    "log_relay/internal/config"

    "log_relay/internal/handlers"
    "log_relay/internal/handlers/auth"
    "log_relay/internal/handlers/list"
    "log_relay/internal/handlers/contact"
    "log_relay/internal/handlers/messages"
    "log_relay/internal/handlers/user"

    _ "log_relay/docs"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
    _ = godotenv.Load() 
    config.CheckRequiredEnvVars()

    env := os.Getenv("GO_ENV")
    if env == "prod" {
        log.Println("Running in Production mode")
        gin.SetMode(gin.ReleaseMode)
    } else {
        log.Println("Running in Development mode")
        gin.SetMode(gin.DebugMode)
    }

    dbHost := os.Getenv("POSTGRESQL_HOST")
    dbPort := os.Getenv("POSTGRESQL_PORT")
    dbPassword := os.Getenv("POSTGRESQL_PASS")
    dbUser, dbName := os.Getenv("POSTGRESQL_USER"), "log_relay"
    db, err := database.ConnectToDB(dbHost, dbUser, dbPassword, dbName, dbPort)
    if err != nil {
        log.Fatalln("Failed to connect to database:", err)
    }
    sqlDB, err := db.DB()
    if err != nil {
        log.Fatalln("Failed to get generic database object:", err)
    }
    if err := sqlDB.Ping(); err != nil {
        log.Fatalln("Database unreachable:", err)
    }
    defer sqlDB.Close()
    log.Println("--Database connection verified--")


    err = db.AutoMigrate(&models.User{}, &models.PasswordReset{}, &models.List{}, &models.Message{}, &models.Contact{}, &models.EmailChangeRequest{},)
    if err != nil {
        log.Fatalf("Migration failed: %v", err)
    }
    log.Println("--Database migration successful-")
    log.Println("--Connected---------------------")

    r := gin.New()

    r.SetTrustedProxies([]string{"127.0.0.1", "::1"})

    r.Use(gin.Recovery())

    if env != "prod" {
        r.Use(gin.Logger())
    }

    
    r.Use(func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")
        allowedOrigins := []string{
            os.Getenv("FRONTEND_URL"),
            os.Getenv("FRONTEND_URL2"),
        }
        for _, o := range allowedOrigins {
            if o == origin {
                c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
                c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
                break
            }
        }
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }
        c.Next()
    })
    
    if env == "dev" {
        r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    }

    r.GET("/status", func(c *gin.Context) {handlers.Ping(c, db) })
    r.POST("/api/user/register", func(c *gin.Context) {auth.CreateUser(c, db) })
    r.POST("/api/user/login", func(c *gin.Context) {auth.LoginUser(c, db) })
    r.POST("/api/user/forgot-password", func(c *gin.Context) {auth.ForgotPassword(c, db) })
    r.POST("/api/user/change-password/:token", func(c *gin.Context) {auth.ResetPassword(c, db) })

    r.POST("/api/subscriber/signup/:list_id", func(c *gin.Context) { contact.ContactSubscribe(c, db) })
    r.GET("/api/subscriber/signup/confirm", func(c *gin.Context) { contact.ContactSubscribeConfirm(c, db) })
    r.DELETE("/api/subscriber/remove/:list_id", func(c *gin.Context) { contact.ContactUnSubscribe(c, db) })
    
    r.GET("/api/list/:list_id", func(c *gin.Context) { list.GetListPublicName(c, db )})
    r.PUT("/api/user/change/email/confirm", func(c*gin.Context) {user.ChangeEmailConfirm(c, db) })

    protected := r.Group("/api")
    protected.Use(middleware.AuthMiddleware())
    {
        protected.GET("/user/me", func(c *gin.Context) { auth.GetMe(c, db) })
        protected.POST("/user/logout", func(c *gin.Context) { auth.LogOut(c, db) })
        
        protected.POST("/user/cycle-token", func(c *gin.Context) { auth.CycleToken(c, db) })
        protected.PUT("/user/change/username", func(c*gin.Context) {user.ChangeUsername(c, db) })
        protected.PATCH("/user/change/email", func(c*gin.Context) {user.ChangeEmail(c, db) })
        protected.DELETE("/user/account/delete", func(c*gin.Context) {user.DeleteAccount(c, db) })

        protected.POST("/list/create", func(c *gin.Context) { list.CreateList(c, db) })
        protected.DELETE("/list/delete/:id", func(c *gin.Context) { list.DeleteList(c, db) })
        protected.GET("/list/index", func(c *gin.Context) { list.IndexList(c, db) })
        protected.GET("/list/detail/:id", func(c *gin.Context) { list.GetListDetail(c, db) })
        
        protected.POST("/mail/send/:list_id", func(c *gin.Context) { messages.SendMailingListMessage(c, db) })
    }

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("Server starting on port %s in %s mode...", port, env)
    r.Run(":" + port)

}
