package config

import (
	"log"
	"os"
)

func CheckRequiredEnvVars() {
	required := []string{
		"GO_ENV",
		"POSTGRESQL_HOST",
		"POSTGRESQL_PASS",
		"POSTGRESQL_USER",
		"POSTGRESQL_PORT",
		"FRONTEND_URL",
		"ADMIN_EMAIL",
		"JWT_SECRET",
		// "AWS_ACCESS_KEY_ID",
		// "AWS_SECRET_ACCESS_KEY",
		// "AWS_DEFAULT_REGION",
		"SENDER_ADDRESS",
		"PORT",
	}
	for _, v := range required {
		if os.Getenv(v) == "" {
			log.Fatalf("CRITICAL CONFIG ERROR: Environment variable '%s' is missing. Server cannot start.", v)
		}
	}
	
	log.Println("Environment validation successful: All variables present.")
}