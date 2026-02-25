package inventory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRawShopeeAPIError(t *testing.T) {
	assert.Equal(t, "E1001: invalid token", rawShopeeAPIError("E1001", "invalid token"))
	assert.Equal(t, "invalid token", rawShopeeAPIError("", "invalid token"))
	assert.Equal(t, "E1001", rawShopeeAPIError("E1001", ""))
	assert.Equal(t, "", rawShopeeAPIError("", ""))
}

func TestRawLazadaAPIError(t *testing.T) {
	assert.Equal(t, "code=1000: Invalid seller sku", rawLazadaAPIError("1000", "Invalid seller sku"))
	assert.Equal(t, "Invalid seller sku", rawLazadaAPIError("", "Invalid seller sku"))
	assert.Equal(t, "1000", rawLazadaAPIError("1000", ""))
	assert.Equal(t, "", rawLazadaAPIError("", ""))
}

func TestRawTiktokAPIError(t *testing.T) {
	assert.Equal(t, "code=50001: invalid parameter", rawTiktokAPIError(50001, "invalid parameter"))
	assert.Equal(t, "code=50001", rawTiktokAPIError(50001, ""))
	assert.Equal(t, "code=0", rawTiktokAPIError(0, ""))
}
