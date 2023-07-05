package main

import "os"

func checkConfig() {
	requiredConfig := []string{"JWT_SECRET", "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_BUCKET_NAME", "AWS_REGION"}

	// Loop over reqired config and check if they are set, and not ""
	for _, v := range requiredConfig {
		if os.Getenv(v) == "" {
			panic("Missing required config: " + v)
		}
	}
}

func DB_PATH() string {
	// Fall back to default tbd.db if not set
	if os.Getenv("DB_PATH") == "" {
		return "tbd.db"
	}
	return os.Getenv("DB_PATH")
}
