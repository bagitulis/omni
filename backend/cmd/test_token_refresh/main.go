//go:build tools
// +build tools

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services"
	"github.com/omni/backend/internal/services/platform"
	"github.com/omni/backend/internal/utils"
)

func main() {
	// Load environment
	tenantID := "yumna_bertigamart"
	testPlatform := "shopee"

	// Setup basic config
	os.Setenv("DB_DRIVER", "postgres")
	os.Setenv("PG_HOST", "localhost")
	os.Setenv("PG_PORT", "5432")
	os.Setenv("PG_USER", "omni")
	os.Setenv("PG_PASSWORD", "omnipassword")
	os.Setenv("PG_DATABASE", "omni_main")

	ctx := context.Background()

	// Connect to system database
	systemDB, err := config.GetSystemDB("./data")
	if err != nil {
		log.Fatalf("Failed to connect to system DB: %v", err)
	}
	log.Println("Connected to system database")

	// Setup encryption
	encryptionKey := os.Getenv("ENCRYPTION_KEY")
	if encryptionKey == "" {
		log.Fatal("ENCRYPTION_KEY not set")
	}
	encryption, err := utils.NewEncryptionService(encryptionKey)
	if err != nil {
		log.Fatalf("Failed to setup encryption: %v", err)
	}

	// Setup global config repo
	globalConfigRepo := repositories.NewGlobalConfigRepository(systemDB)

	// Setup token manager
	tokenManager := services.NewTokenManager(globalConfigRepo, encryption, "./data")

	// Register token refresh adapter
	tokenRefreshAdapter := services.NewTokenRefreshAdapter(tokenManager)
	platform.RegisterTokenRefreshService(tokenRefreshAdapter)
	log.Println("Token refresh service registered")

	// Test 1: Get token status
	fmt.Println("\n=== TEST 1: Get Token Status ===")
	status, err := tokenManager.GetTokenStatus(ctx, tenantID, testPlatform)
	if err != nil {
		log.Printf("Error getting token status: %v", err)
	} else {
		fmt.Printf("Platform: %s\n", status.Platform)
		fmt.Printf("Is Valid: %v\n", status.IsValid)
		fmt.Printf("Needs Refresh: %v\n", status.NeedsRefresh)
		fmt.Printf("Expires At: %v\n", status.ExpiresAt.Format(time.RFC3339))
		fmt.Printf("Refresh Token Expires: %v\n", status.RefreshTokenExpires.Format(time.RFC3339))
	}

	// Test 2: Initialize platform coordination service
	fmt.Println("\n=== TEST 2: Initialize Platform Coordination Service ===")
	coordService := platform.GetPlatformCoordinationService(tenantID)
	if err := coordService.InitializePlatforms(ctx); err != nil {
		log.Printf("Error initializing platforms: %v", err)
	} else {
		log.Println("Platforms initialized successfully")
	}

	// Test 3: Get Shopee client and check if token refresher is set
	fmt.Println("\n=== TEST 3: Test Shopee GetOrderList (this should trigger auto-refresh) ===")
	shopeeClient := coordService.GetShopeeClient()
	if shopeeClient == nil {
		log.Println("Shopee client not available")
	} else {
		if shopeeClient.IsInitialized() {
			log.Println("Shopee client initialized, fetching orders...")
			orders, err := shopeeClient.GetOrderList(ctx, "READY_TO_SHIP", 7)
			if err != nil {
				log.Printf("Error getting orders (expected - token might be expired): %v", err)
			} else {
				fmt.Printf("Got %d orders\n", len(orders))
			}
		} else {
			log.Println("Shopee client not properly configured")
		}
	}

	// Test 4: Manual token refresh
	fmt.Println("\n=== TEST 4: Manual Token Refresh ===")
	newToken, err := tokenManager.RefreshShopeeToken(ctx, tenantID)
	if err != nil {
		log.Printf("Error refreshing token: %v", err)
	} else {
		fmt.Printf("New Token: %s...\n", newToken.AccessToken[:20])
		fmt.Printf("New Expires At: %v\n", newToken.ExpiresAt.Format(time.RFC3339))
	}

	fmt.Println("\n=== DONE ===")
}
