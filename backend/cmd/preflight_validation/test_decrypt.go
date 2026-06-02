package main

import (
	"fmt"
	"os"

	"github.com/omni/backend/internal/utils"
)

func main() {
	encKey := os.Getenv("ENCRYPTION_KEY")
	if encKey == "" {
		fmt.Println("ERROR: ENCRYPTION_KEY not set")
		os.Exit(1)
	}

	enc, err := utils.NewEncryptionService(encKey)
	if err != nil {
		fmt.Printf("ERROR: Failed to create encryption service: %v\n", err)
		os.Exit(1)
	}

	// Encrypted values from production database (bertigamart tenant)
	testValues := map[string]string{
		"lazada_accessToken":  "gAAAAABqHm6FjaZef004ShF5B-CkTdzH6I87sYWihOl518-93OloS9dqr0pMIthgodtbk5pVvouyxG7Czjd8KtRnP2d9GTgt1vF2UmhU2CPwoBvlIUMFUMt1QeYd1WyPO_5wt7VZC3ItHEyTeGFv4BMgITTS7MeixpiBPZ_UE60jiYbewwwz64A=",
		"lazada_refreshToken": "gAAAAABqHm6FjnI2pUGIR2qOcbfecG2qge2Lm15YulGlfJuLVsL5yv1t7-eEmDhZWXUuew2e0QyyhfZqPs-ZOt9WhwT8zhrjJjUAkGhZ8NY3-ch79w_uyiCFejhSULIzIJu1rU43cD-onsX2uo6ENboYVGqUgygYkEEgHXsDjaZPQyinDlAMH1I=",
		"shopee_accessToken":  "gAAAAABqGCDdkWgGoLJDUarcZyUZkUH2A72gH4EC8PoZbIcTfcykt7zS2dDvyYrCMpdvr1z4ugSkTGr1tNhBrYRip13hjBWSoQfJnjFIvyxVU8Tv3Ky5zXN8326gcKdG41k1OIVnwghs2FddZeh0EdX9WhRUpjmFPzRDUrAIkta5IejEXW86FpVO5EO-l06lFwMEQnqdwJyI5X79FySrNJCeedb7-Fe-bg==",
		"shopee_refreshToken": "gAAAAABqGCDd4kl80xZeIxlFogT4Hp6qYPhfjWyX7AAbOkW71EAXlJzGl0mnpLioFCfaAtKdvRHPaBvOebcnsrWUDF1Dv-kCoqUnwlWUkb6T9bA0ctIFyLy5VxN6b_nIIBoS-Xdrl65N4o-55FH4jDCz0GLXB_IaSsMfyGBN1t-bMvmxH3Det5NDDTicCcKXWOvv7lEo2af-shfb16pUCBQTL6Fr4OS0PA==",
		"tiktok_accessToken":  "gAAAAABqHm6GygeLKFxsxodi5pG9MyTD690GHURsrfZkGsNfpX-K3Sa8a-Ft8EXYkRnmWj_gaeyt7jBjmaNpX6hqhRuiQoubL5f-KQmZtflZt9vTb4zseWL-GHM8xnNRzVtpUmFpZXS2I_lqlOw6-iZEg1Ti4VGQ5ZdY8_OThh74rgGBFrQzOCMiZ82xsQ_Fi8UZKQzswTOB0gpY69JMVHDqTqrcr4NHyKqOeYRGoI3171q3zKO4siiE9_LfDKF72r9BnIKONrTsGJ1gl0KhFonN9hnVkr09LzQ2BPOBTo0jTFPKWJrMy9SGJr-xZjAtcBg19nvf_rnT",
		"tiktok_refreshToken": "gAAAAABqHm6GgO7w12u82LwPFG5KvyNHCZ6PXMZISKQg4WSiZGz4-PxacYhkKjtaBbs15fpSD994r83_zzauk9M73xC3jIzPWNKQgfg8a6dJyaN9vZ3olInbyB9ZxwkQ0TaLidWEjJBvhdigNwwiHF3k820a2380784mocnCL2LwSOtyk3DI7ko=",
	}

	passCount := 0
	failCount := 0

	for name, value := range testValues {
		_, err := enc.Decrypt(value)
		if err != nil {
			fmt.Printf("FAIL: %s - %v\n", name, err)
			failCount++
		} else {
			fmt.Printf("PASS: %s\n", name)
			passCount++
		}
	}

	fmt.Printf("\n=== RESULTS ===\n")
	fmt.Printf("Total: %d\n", len(testValues))
	fmt.Printf("Pass:  %d\n", passCount)
	fmt.Printf("Fail:  %d\n", failCount)

	if failCount == len(testValues) {
		fmt.Println("\n*** CRITICAL: ALL values failed - Fernet key may have changed! ***")
		os.Exit(2)
	} else if failCount > 0 {
		fmt.Println("\n*** WARNING: Some values failed decryption ***")
		os.Exit(1)
	} else {
		fmt.Println("\n*** ALL VALUES PASS - encryption key is valid ***")
	}
}
