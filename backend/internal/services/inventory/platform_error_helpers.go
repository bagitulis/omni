package inventory

import "fmt"

func rawShopeeAPIError(code string, message string) string {
	if code == "" {
		return message
	}
	if message == "" {
		return code
	}

	return fmt.Sprintf("%s: %s", code, message)
}

func rawLazadaAPIError(code string, message string) string {
	if code == "" {
		return message
	}
	if message == "" {
		return code
	}

	return fmt.Sprintf("code=%s: %s", code, message)
}

func rawTiktokAPIError(code int, message string) string {
	if message == "" {
		return fmt.Sprintf("code=%d", code)
	}

	return fmt.Sprintf("code=%d: %s", code, message)
}
