package main

import (
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"

    "log_relay/internal/database"
    "log_relay/internal/models"
    "log_relay/internal/handlers"
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

    err = db.AutoMigrate(&models.Sender{}, &models.User{}, &models.Message{}, &models.PasswordReset{},)
    if err != nil {
        log.Fatalf("Migration failed: %v", err)
    }
    log.Println("--Database migration successful-")
    log.Println("--Connected---------------------")
    r := gin.Default()
    
    // Pass db to handlers 
    r.GET("/status", func(c *gin.Context) {handlers.Ping(c, db)})
    r.POST("/user/register", func(c *gin.Context) {handlers.CreateUser(c, db)})
    r.POST("/user/login", func(c *gin.Context) {handlers.LoginUser(c, db)})
    r.POST("/user/cycle-token", func(c *gin.Context) {handlers.CycleToken(c, db)})
    r.POST("/user/reset-password", func(c *gin.Context) {handlers.ForgotPassword(c, db)})
    r.POST("/user/change-password/:token", func(c *gin.Context) {handlers.ResetPassword(c, db)})
    r.POST("/user/logout", func(c *gin.Context) {handlers.LogOut(c, db)})
    
    r.Run() // Default is port 8080

}