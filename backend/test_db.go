// test_db.go
package main

import (
    "context"
    "fmt"
    "log"
    "strings"
    "time"

    "github.com/andikatampubolon10/hris-backend/internal/config"
    "github.com/andikatampubolon10/hris-backend/pkg/database"
)

func main() {
    separator := strings.Repeat("=", 60)
    
    fmt.Println(separator)
    fmt.Println("🔍 TESTING MONGODB ATLAS CONNECTION")
    fmt.Println(separator + "\n")

    // Load configuration
    fmt.Println("📋 Step 1: Loading configuration...")
    cfg := config.LoadConfig()

    // Display connection info (without password)
    fmt.Printf("   ✓ Database Name: %s\n", cfg.DatabaseName)
    fmt.Printf("   ✓ Environment: %s\n", cfg.Environment)
    fmt.Println()

    // Connect to MongoDB
    fmt.Println("🔌 Step 2: Connecting to MongoDB Atlas...")
    mongodb, err := database.NewMongoDB(cfg.MongoURI, cfg.DatabaseName)
    if err != nil {
        log.Fatal("   ❌ Connection failed:", err)
    }
    defer mongodb.Disconnect()

    fmt.Println("   ✅ Connection successful!")
    fmt.Println()

    // Test operations
    fmt.Println("🧪 Step 3: Testing database operations...")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // List databases
    databases, err := mongodb.Client.ListDatabaseNames(ctx, map[string]interface{}{})
    if err != nil {
        log.Fatal("   ❌ Failed to list databases:", err)
    }

    fmt.Println("   📊 Available databases:")
    for _, db := range databases {
        fmt.Printf("      • %s\n", db)
    }
    fmt.Println()

    // List collections in our database
    collections, err := mongodb.Database.ListCollectionNames(ctx, map[string]interface{}{})
    if err != nil {
        log.Fatal("   ❌ Failed to list collections:", err)
    }

    fmt.Printf("   📁 Collections in '%s':\n", cfg.DatabaseName)
    if len(collections) == 0 {
        fmt.Println("      (No collections yet - will be created automatically)")
    } else {
        for _, coll := range collections {
            fmt.Printf("      • %s\n", coll)
        }
    }
    fmt.Println()

    // Test ping
    fmt.Println("🏓 Step 4: Testing ping...")
    err = mongodb.Client.Ping(ctx, nil)
    if err != nil {
        log.Fatal("   ❌ Ping failed:", err)
    }
    fmt.Println("   ✅ Ping successful!")
    fmt.Println()

    fmt.Println(separator)
    fmt.Println("🎉 ALL TESTS PASSED!")
    fmt.Println(separator)
    fmt.Println("\n✅ Database is ready to use!")
}