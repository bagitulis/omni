package services

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/omni/backend/internal/utils"
)

func buildLegacyBundles(tenantID string, rows []inventoryRowData) ([]legacyCredentialBundle, []string) {
	keyValueRows := map[string][]inventoryRowData{}
	structuredBundles := map[string]legacyCredentialBundle{}
	abortReasons := make([]string, 0)
	for _, row := range rows {
		shape := detectRowShape(row)
		if shape == rowShapeUnknown {
			abortReasons = append(abortReasons, "unknown_row_shape")
			continue
		}
		if len(missingRequiredFields(shape, row)) > 0 {
			abortReasons = append(abortReasons, "incomplete_source")
			continue
		}
		if reason := validateEncryptedValues(row); reason != "" {
			abortReasons = append(abortReasons, reason)
		}
		platform := strings.TrimSpace(asString(row["platform"]))
		if shape == rowShapeKeyValue {
			keyValueRows[tenantID+"|"+platform] = append(keyValueRows[tenantID+"|"+platform], row)
			continue
		}
		bundle := structuredRowToBundle(tenantID, row)
		key := bundle.TenantID + "|" + bundle.Platform + "|" + bundle.StoreIdentifier
		structuredBundles[key] = mergeLegacyBundle(structuredBundles[key], bundle)
	}
for _, rows := range keyValueRows {
	bundle := keyValueRowsToBundle(tenantID, rows)
	if bundle.StoreIdentifier == "" && (bundle.AccessToken != "" || bundle.RefreshToken != "" || bundle.ShopCipher != "") {
		abortReasons = append(abortReasons, "missing_store_identity")
	}
	key := bundle.TenantID + "|" + bundle.Platform + "|" + bundle.StoreIdentifier
	if bundle.StoreIdentifier == "" {
		key = bundle.TenantID + "|" + bundle.Platform + "|__app_only__"
	}
	structuredBundles[key] = mergeLegacyBundle(structuredBundles[key], bundle)
}
	bundles := make([]legacyCredentialBundle, 0, len(structuredBundles))
	for _, bundle := range structuredBundles {
		bundles = append(bundles, bundle)
	}
	sort.Slice(bundles, func(i, j int) bool {
		return bundles[i].Platform+bundles[i].StoreIdentifier < bundles[j].Platform+bundles[j].StoreIdentifier
	})
	return bundles, abortReasons
}

func structuredRowToBundle(tenantID string, row inventoryRowData) legacyCredentialBundle {
	return legacyCredentialBundle{
		TenantID:        tenantID,
		Platform:        strings.TrimSpace(asString(row["platform"])),
		StoreIdentifier: firstNonEmpty(asString(row["shop_id"]), asString(row["shop_name"])),
		StoreName:       asString(row["shop_name"]),
		Region:          "id",
		AccessToken:     asString(row["access_token"]),
		RefreshToken:    asString(row["refresh_token"]),
		ShopCipher:      asString(row["shop_cipher"]),
		TokenExpiry:     parseLegacyTime(row["token_expires_at"]),
	}
}

func keyValueRowsToBundle(tenantID string, rows []inventoryRowData) legacyCredentialBundle {
	bundle := legacyCredentialBundle{TenantID: tenantID, Region: "id"}
	values := map[string]string{}
	for _, row := range rows {
		bundle.Platform = strings.TrimSpace(asString(row["platform"]))
		key := strings.TrimSpace(asString(row["config_key"]))
		values[key] = strings.TrimSpace(asString(row["config_value"]))
		bundle.SourceKeys = append(bundle.SourceKeys, key)
	}
	bundle.StoreIdentifier = firstNonEmpty(values["shopId"], values["shop_id"], values["shopName"], values["shop_name"])
	bundle.StoreName = firstNonEmpty(values["shopName"], values["shop_name"])
	bundle.AccessToken = firstNonEmpty(values["accessToken"], values["access_token"])
	bundle.RefreshToken = firstNonEmpty(values["refreshToken"], values["refresh_token"])
	bundle.ShopCipher = firstNonEmpty(values["shopCipher"], values["shop_cipher"])
	bundle.AppKey = values["appKey"]
	bundle.AppSecret = values["appSecret"]
	bundle.PartnerKey = values["partnerKey"]
	bundle.PartnerID, _ = strconv.ParseInt(values["partnerId"], 10, 64)
	bundle.TokenExpiry, _ = strconv.ParseInt(firstNonEmpty(values["tokenExpiry"], values["expiresAt"]), 10, 64)
	bundle.Region = normalizeRegion(firstNonEmpty(values["region"], values["country"], "id"))
	return bundle
}

func validateEncryptedValues(row inventoryRowData) string {
	isEncrypted := strings.EqualFold(asString(row["is_encrypted"]), "true")
	value := strings.TrimSpace(asString(row["config_value"]))
	if isEncrypted && value != "" && !utils.IsEncrypted(value) {
		return "unreadable_encrypted_value"
	}
	return ""
}

func mergeLegacyBundle(left, right legacyCredentialBundle) legacyCredentialBundle {
	if left.TenantID == "" {
		return right
	}
	left.StoreName = firstNonEmpty(left.StoreName, right.StoreName)
	left.Region = firstNonEmpty(left.Region, right.Region)
	left.AccessToken = firstNonEmpty(left.AccessToken, right.AccessToken)
	left.RefreshToken = firstNonEmpty(left.RefreshToken, right.RefreshToken)
	left.ShopCipher = firstNonEmpty(left.ShopCipher, right.ShopCipher)
	left.AppKey = firstNonEmpty(left.AppKey, right.AppKey)
	left.AppSecret = firstNonEmpty(left.AppSecret, right.AppSecret)
	left.PartnerKey = firstNonEmpty(left.PartnerKey, right.PartnerKey)
	if left.PartnerID == 0 {
		left.PartnerID = right.PartnerID
	}
	if left.TokenExpiry == 0 {
		left.TokenExpiry = right.TokenExpiry
	}
	left.SourceKeys = append(left.SourceKeys, right.SourceKeys...)
	return left
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func parseLegacyTime(value any) int64 {
	switch v := value.(type) {
	case time.Time:
		return v.UnixMilli()
	case string:
		parsed, _ := time.Parse(time.RFC3339, v)
		return parsed.UnixMilli()
	default:
		return 0
	}
}
