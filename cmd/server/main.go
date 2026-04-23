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

    _ "log_relay/docs"
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
    _ = godotenv.Load() 

    config.CheckRequiredEnvVars()


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

    err = db.AutoMigrate(&models.User{}, &models.PasswordReset{}, &models.List{}, &models.Message{}, &models.Contact{},)
    if err != nil {
        log.Fatalf("Migration failed: %v", err)
    }
    log.Println("--Database migration successful-")
    log.Println("--Connected---------------------")

    r := gin.Default()

    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    r.Use(func(c *gin.Context) {
        frontendURL := os.Getenv("FRONTEND_URL")
		c.Writer.Header().Set("Access-Control-Allow-Origin", frontendURL)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
        c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})
    


    r.GET("/status", func(c *gin.Context) {handlers.Ping(c, db)})
    r.POST("/api/user/register", func(c *gin.Context) {auth.CreateUser(c, db)})
    r.POST("/api/user/login", func(c *gin.Context) {auth.LoginUser(c, db)})
    r.POST("/api/user/forgot-password", func(c *gin.Context) {auth.ForgotPassword(c, db)})
    r.POST("/api/user/change-password/:token", func(c *gin.Context) {auth.ResetPassword(c, db)})

    r.POST("/api/subscriber/signup/:list_id", func(c *gin.Context) { contact.ContactSubscribe(c, db)})
    r.GET("/api/subscriber/signup/confirm", func(c *gin.Context) { contact.ContactSubscribeConfirm(c, db)})
    // implement by adding a link to the front end to sub from mailing list
    r.DELETE("/api/subscriber/remove/:list_id/confirm/:token", func(c *gin.Context) { contact.ContactUnSubscribe(c, db)})


    protected := r.Group("/api")
    protected.Use(middleware.AuthMiddleware())
    {
        protected.GET("/user/me", func(c *gin.Context) { auth.GetMe(c, db) })
        protected.POST("/user/cycle-token", func(c *gin.Context) { auth.CycleToken(c, db) })
        protected.POST("/user/logout", func(c *gin.Context) { auth.LogOut(c, db) })

        protected.POST("/list/create", func(c *gin.Context) { list.CreateList(c, db) })
        protected.DELETE("/list/delete/:id", func(c *gin.Context) { list.DeleteList(c, db) })
        protected.GET("/list/index", func(c *gin.Context) { list.IndexList(c, db) })
        protected.GET("/list/detail/:id", func(c *gin.Context) { list.GetListDetail(c, db) })
        
        protected.POST("/list/send-mail/:list_id", func(c *gin.Context) { messages.SendMailingListMessage(c, db)})
    }
    r.Run() // Default is port 8080

}
