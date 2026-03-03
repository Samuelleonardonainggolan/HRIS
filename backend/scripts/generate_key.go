package main

import (
	"fmt"
	"github.com/andikatampubolon10/hris-backend/pkg/auth"
)

func main() {
	fmt.Println(auth.GenerateRandomKey())
}
