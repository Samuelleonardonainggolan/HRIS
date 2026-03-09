// test_password.go
package main

import (
	"fmt"
	"log"

	"github.com/andikatampubolon10/hris-backend/pkg/auth"
)

func main() {
	password := "password123"
	
	// Hash password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}
	
	fmt.Println("Password:", password)
	fmt.Println("Hashed:", hashedPassword)
	
	// Check if password matches
	matches := auth.CheckPasswordHash(password, hashedPassword)
	fmt.Println("Password matches:", matches)
}