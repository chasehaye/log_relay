package services

import "os"

var envIsProduction bool
func init() {
    envIsProduction = os.Getenv("GO_ENV") == "production"
}

func IsProduction() bool {
    return envIsProduction
}