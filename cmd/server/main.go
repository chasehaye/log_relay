package main

import (
    "log"
    "os"
    // comment out for dev
    // "net/http"

    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"

    "log_relay/internal/database"
    "log_relay/internal/models"
    "log_relay/internal/handlers"
    "log_relay/internal/middleware"
)

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
    dbPassword := os.Getenv("POSTGRESQL_PASS")
    dbUser, dbName := os.Getenv("POSTGRESQL_USER"), "log_relay"

    db, err := database.ConnectToDB(dbUser, dbPassword, dbName)
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

    err = db.AutoMigrate(&models.User{}, &models.PasswordReset{}, &models.List{}, &models.Message{}, &models.Contact{},)
    if err != nil {
        log.Fatalf("Migration failed: %v", err)
    }
    log.Println("--Database migration successful-")
    log.Println("--Connected---------------------")
    r := gin.Default()

    // comment out for dev
    // r.Use(func(c *gin.Context) {
    //     frontendURL := os.Getenv("FRONTEND_URL")
	// 	c.Writer.Header().Set("Access-Control-Allow-Origin", frontendURL)
	// 	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	// 	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	// 	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
    //     c.Writer.Header().Set("Access-Control-Max-Age", "86400")

	// 	if c.Request.Method == "OPTIONS" {
	// 		c.AbortWithStatus(http.StatusNoContent)
	// 		return
	// 	}
	// 	c.Next()
	// })
    // comment out for dev^
    
    // handlers
    r.GET("/status", func(c *gin.Context) {handlers.Ping(c, db)})
    r.POST("/api/user/register", func(c *gin.Context) {handlers.CreateUser(c, db)})
    r.POST("/api/user/login", func(c *gin.Context) {handlers.LoginUser(c, db)})
    r.POST("/api/user/forgot-password", func(c *gin.Context) {handlers.ForgotPassword(c, db)})
    r.POST("/api/user/change-password/:token", func(c *gin.Context) {handlers.ResetPassword(c, db)})

    auth := r.Group("/api")
    auth.Use(middleware.AuthMiddleware())
    {
        auth.POST("/user/cycle-token", func(c *gin.Context) { handlers.CycleToken(c, db) })
        auth.POST("/user/logout", func(c *gin.Context) { handlers.LogOut(c, db) })
        auth.POST("/list/create", func(c *gin.Context) { handlers.CreateList(c, db) })
    }

    r.Run() // Default is port 8080

}