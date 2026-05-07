// @title           Log Relay API
// @version         1.0
// @description     API Server for Log Relay Project.
// @host            localhost:8080
// @BasePath        /
package main

import (
    "log"
    "os"


    "github.com/joho/godotenv"

    "log_relay/internal/database"
    "log_relay/internal/models"
    "log_relay/internal/config"
    "log_relay/internal/migrations"
    "log_relay/internal/router"
)

func main() {
    // ------------------ ENV VALIDATION ------------------
    _ = godotenv.Load() 
    cfg := config.CheckRequiredEnvVarsAndLoad()

    // ------------------ DATABASE ------------------
    db, err := database.ConnectToDB(cfg.DBHost, cfg.DBUser, cfg.DBPass, "log_relay", cfg.DBPort)
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
    if err := migrations.FixLegacyData(db); err != nil {
        log.Fatalf("Legacy migration failed: %v", err)
    }

    // ------------------ Migrate ------------------
    err = db.AutoMigrate(
        &models.User{}, 
        &models.PasswordReset{}, 
        &models.List{}, 
        &models.Message{}, 
        &models.Contact{}, 
        &models.EmailChangeRequest{},
    )
    if err != nil {
        log.Fatalf("Migration failed: %v", err)
    }

    log.Println("--Database migration successful-")
    log.Println("--Connected---------------------")

    // ------------------ START SERVER ------------------
    r := router.Setup(db)
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("Server starting on port %s in %s mode...", cfg.Port, cfg.Env)
    r.Run(":" + port)

}