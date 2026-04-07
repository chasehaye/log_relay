package services

import "os"

var envIsProduction bool

func init() {
    env := os.Getenv("GO_ENV")

    switch env {
    case "prod", "production":
        envIsProduction = true
    case "dev", "development":
        envIsProduction = false
    default:
        envIsProduction = false
    }
}

func IsProduction() bool {
    return envIsProduction
}