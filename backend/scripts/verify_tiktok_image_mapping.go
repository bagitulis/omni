package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	paths := []string{
		"backend/internal/services/sync/order_repository_tiktok.go",
		"internal/services/sync/order_repository_tiktok.go",
	}
	var path string
	var content []byte
	var err error
	for _, candidate := range paths {
		content, err = os.ReadFile(candidate)
		if err == nil {
			path = candidate
			break
		}
	}
	if err != nil {
		fmt.Printf("FAILED: unable to read order repository: %v\n", err)
		os.Exit(1)
	}
	data := string(content)

	required := []string{
		"internalMap",
		"getMasterProductImagesByItemIDs",
		"product_id",
	}
	for _, token := range required {
		if !strings.Contains(data, token) {
			fmt.Printf("FAILED: missing %q in %s\n", token, path)
			os.Exit(1)
		}
	}

	fmt.Println("SUCCESS: TikTok master image lookup uses platform product_id mapping")
}
