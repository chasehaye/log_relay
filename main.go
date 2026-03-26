package main

import (
    "log_relay/handlers"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    r.GET("/ping", handlers.Ping)

    r.Run() // listens on :8080 by default
}