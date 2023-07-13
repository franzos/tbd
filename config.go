package main

import (
	"fmt"
	"os"
)

func checkConfig() {
	requiredConfig := []string{"JWT_SECRET", "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_BUCKET_NAME", "AWS_REGION", "DOMAIN", "PGP_PASSPHRASE"}

	// Loop over reqired config and check if they are set, and not ""
	for _, v := range requiredConfig {
		if os.Getenv(v) == "" {
			panic("Missing required config: " + v)
		}
	}

	file1 := "./auth_model.conf"
	file2 := "./policy.csv"

	// Check if file1 exists
	_, err := os.Stat(file1)
	if err != nil {
		if os.IsNotExist(err) {
			panic("Missing required config: " + file1)
		} else {
			fmt.Printf("Error checking file %s: %v\n", file1, err)
		}
	} else {
		fmt.Printf("Check: File %s exists\n", file1)
	}

	// Check if file2 exists
	_, err = os.Stat(file2)
	if err != nil {
		if os.IsNotExist(err) {
			panic("Missing required config: " + file2)
		} else {
			panic("Error checking file " + file2)
		}
	} else {
		fmt.Printf("Check: File %s exists\n", file2)
	}
}

func DB_PATH() string {
	// Fall back to default tbd.db if not set
	if os.Getenv("DB_PATH") == "" {
		return "tbd.db"
	}
	return os.Getenv("DB_PATH")
}
