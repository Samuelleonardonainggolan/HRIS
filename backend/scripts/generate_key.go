//go:build generate_key

package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func main() {
	// Generate 32 random bytes and encode to base64
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	fmt.Println(base64.StdEncoding.EncodeToString(b))
}
