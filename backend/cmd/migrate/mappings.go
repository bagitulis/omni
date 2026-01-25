package main

// getTableMappings returns all table mappings for migration
func getTableMappings() []TableMapping {
	return []TableMapping{
		getUserMapping(),
		getPlatformConfigMapping(),
		getShopeeOrderMapping(),
		getShopeeOrderItemMapping(),
		getLazadaOrderMapping(),
		getTiktokOrderMapping(),
		getShopeeProductMapping(),
		getInventoryRecordMapping(),
		getFilterPreferenceMapping(),
		getWholesaleSettingsMapping(),
		getWebhookLogMapping(),
		getOAuthStateMapping(),
	}
}

func getUserMapping() TableMapping {
	return TableMapping{
		SQLiteTable:   "User",
		PostgresTable: "users",
		Columns: []ColumnMapping{
			{SQLiteCol: "id", PostgresCol: "id"},
			{SQLiteCol: "username", PostgresCol: "username"},
			{SQLiteCol: "email", PostgresCol: "email"},
			{SQLiteCol: "password", PostgresCol: "password"},
			{SQLiteCol: "role", PostgresCol: "role"},
			{SQLiteCol: "failedLoginAttempts", PostgresCol: "failed_login_attempts"},
			{SQLiteCol: "accountLockedUntil", PostgresCol: "account_locked_until", Transform: transformDateTime},
			{SQLiteCol: "lastFailedLogin", PostgresCol: "last_failed_login", Transform: transformDateTime},
			{SQLiteCol: "createdAt", PostgresCol: "created_at", Transform: transformDateTime},
			{SQLiteCol: "updatedAt", PostgresCol: "updated_at", Transform: transformDateTime},
		},
	}
}

func getPlatformConfigMapping() TableMapping {
	return TableMapping{
		SQLiteTable:   "PlatformConfig",
		PostgresTable: "platform_configs",
		Columns: []ColumnMapping{
			{SQLiteCol: "id", PostgresCol: "id"},
			{SQLiteCol: "platform", PostgresCol: "platform"},
			{SQLiteCol: "configKey", PostgresCol: "config_key"},
			{SQLiteCol: "configValue", PostgresCol: "config_value"},
			{SQLiteCol: "dataType", PostgresCol: "data_type"},
			{SQLiteCol: "isEncrypted", PostgresCol: "is_encrypted", Transform: transformBoolean},
			{SQLiteCol: "metadata", PostgresCol: "metadata", Transform: transformJSON},
			{SQLiteCol: "createdAt", PostgresCol: "created_at", Transform: transformDateTime},
			{SQLiteCol: "updatedAt", PostgresCol: "updated_at", Transform: transformDateTime},
		},
	}
}

func getShopeeOrderMapping() TableMapping {
	return TableMapping{
		SQLiteTable:   "ShopeeOrder",
		PostgresTable: "shopee_orders",
		HasTenantID:   true,
		Columns: []ColumnMapping{
			{SQLiteCol: "id", PostgresCol: "id"},
			{SQLiteCol: "tenantId", PostgresCol: "tenant_id"},
			{SQLiteCol: "orderSn", PostgresCol: "order_sn"},
			{SQLiteCol: "shopId", PostgresCol: "shop_id"},
			{SQLiteCol: "orderStatus", PostgresCol: "order_status"},
			{SQLiteCol: "orderTimestamp", PostgresCol: "order_timestamp"},
			{SQLiteCol: "createdAt", PostgresCol: "created_at", Transform: transformDateTime},
			{SQLiteCol: "updatedAt", PostgresCol: "updated_at", Transform: transformDateTime},
		},
	}
}

func getShopeeOrderItemMapping() TableMapping {
	return TableMapping{
		SQLiteTable:   "ShopeeOrderItem",
		PostgresTable: "shopee_order_items",
		Columns: []ColumnMapping{
			{SQLiteCol: "id", PostgresCol: "id"},
			{SQLiteCol: "orderSn", PostgresCol: "order_sn"},
			{SQLiteCol: "itemId", PostgresCol: "item_id"},
			{SQLiteCol: "modelId", PostgresCol: "model_id"},
			{SQLiteCol: "itemName", PostgresCol: "item_name"},
			{SQLiteCol: "modelName", PostgresCol: "model_name"},
			{SQLiteCol: "itemSku", PostgresCol: "item_sku"},
			{SQLiteCol: "modelSku", PostgresCol: "model_sku"},
			{SQLiteCol: "quantity", PostgresCol: "quantity"},
			{SQLiteCol: "price", PostgresCol: "price"},
			{SQLiteCol: "createdAt", PostgresCol: "created_at", Transform: transformDateTime},
			{SQLiteCol: "updatedAt", PostgresCol: "updated_at", Transform: transformDateTime},
		},
	}
}

func getLazadaOrderMapping() TableMapping {
	return TableMapping{
		SQLiteTable:   "LazadaOrder",
		PostgresTable: "lazada_orders",
		HasTenantID:   true,
		Columns: []ColumnMapping{
			{SQLiteCol: "id", PostgresCol: "id"},
			{SQLiteCol: "tenantId", PostgresCol: "tenant_id"},
			{SQLiteCol: "orderSn", PostgresCol: "order_sn"},
			{SQLiteCol: "shopId", PostgresCol: "shop_id"},
			{SQLiteCol: "orderStatus", PostgresCol: "order_status"},
			{SQLiteCol: "createdAt", PostgresCol: "created_at", Transform: transformDateTime},
			{SQLiteCol: "updatedAt", PostgresCol: "updated_at", Transform: transformDateTime},
		},
	}
}

func getTiktokOrderMapping() TableMapping {
	return TableMapping{
		SQLiteTable:   "TiktokOrder",
		PostgresTable: "tiktok_orders",
		HasTenantID:   true,
		Columns: []ColumnMapping{
			{SQLiteCol: "id", PostgresCol: "id"},
			{SQLiteCol: "tenantId", PostgresCol: "tenant_id"},
			{SQLiteCol: "orderSn", PostgresCol: "order_sn"},
			{SQLiteCol: "shopId", PostgresCol: "shop_id"},
			{SQLiteCol: "orderStatus", PostgresCol: "order_status"},
			{SQLiteCol: "createdAt", PostgresCol: "created_at", Transform: transformDateTime},
			{SQLiteCol: "updatedAt", PostgresCol: "updated_at", Transform: transformDateTime},
		},
	}
}

func getShopeeProductMapping() TableMapping {
	return TableMapping{
		SQLiteTable:   "ShopeeProduct",
		PostgresTable: "shopee_products",
		HasTenantID:   true,
		Columns: []ColumnMapping{
			{SQLiteCol: "id", PostgresCol: "id"},
			{SQLiteCol: "tenantId", PostgresCol: "tenant_id"},
			{SQLiteCol: "itemId", PostgresCol: "item_id"},
			{SQLiteCol: "name", PostgresCol: "name"},
			{SQLiteCol: "description", PostgresCol: "description"},
			{SQLiteCol: "status", PostgresCol: "status"},
			{SQLiteCol: "price", PostgresCol: "price"},
			{SQLiteCol: "quantity", PostgresCol: "quantity"},
			{SQLiteCol: "image", PostgresCol: "image"},
			{SQLiteCol: "createdAt", PostgresCol: "created_at", Transform: transformDateTime},
			{SQLiteCol: "updatedAt", PostgresCol: "updated_at", Transform: transformDateTime},
		},
	}
}

func getInventoryRecordMapping() TableMapping {
	return TableMapping{
		SQLiteTable:   "InventoryRecord",
		PostgresTable: "inventory_records",
		HasTenantID:   true,
		Columns: []ColumnMapping{
			{SQLiteCol: "id", PostgresCol: "id"},
			{SQLiteCol: "tenantId", PostgresCol: "tenant_id"},
			{SQLiteCol: "data", PostgresCol: "data", Transform: transformJSON},
			{SQLiteCol: "keyValue", PostgresCol: "key_value"},
			{SQLiteCol: "keyColumnName", PostgresCol: "key_column_name"},
			{SQLiteCol: "createdAt", PostgresCol: "created_at", Transform: transformDateTime},
			{SQLiteCol: "updatedAt", PostgresCol: "updated_at", Transform: transformDateTime},
		},
	}
}

func getFilterPreferenceMapping() TableMapping {
	return TableMapping{
		SQLiteTable:   "FilterPreference",
		PostgresTable: "filter_preferences",
		HasTenantID:   true,
		Columns: []ColumnMapping{
			{SQLiteCol: "id", PostgresCol: "id"},
			{SQLiteCol: "tenantId", PostgresCol: "tenant_id"},
			{SQLiteCol: "platform", PostgresCol: "platform"},
			{SQLiteCol: "tab", PostgresCol: "tab"},
			{SQLiteCol: "visibleColumns", PostgresCol: "visible_columns"},
			{SQLiteCol: "columnFilters", PostgresCol: "column_filters"},
			{SQLiteCol: "searchQuery", PostgresCol: "search_query"},
			{SQLiteCol: "lockedColumns", PostgresCol: "locked_columns"},
			{SQLiteCol: "createdAt", PostgresCol: "created_at", Transform: transformDateTime},
			{SQLiteCol: "updatedAt", PostgresCol: "updated_at", Transform: transformDateTime},
		},
	}
}

func getWholesaleSettingsMapping() TableMapping {
	return TableMapping{
		SQLiteTable:   "WholesaleSettings",
		PostgresTable: "wholesale_settings",
		HasTenantID:   true,
		Columns: []ColumnMapping{
			{SQLiteCol: "id", PostgresCol: "id"},
			{SQLiteCol: "tenantId", PostgresCol: "tenant_id"},
			{SQLiteCol: "platform", PostgresCol: "platform"},
			{SQLiteCol: "adminFee", PostgresCol: "admin_fee"},
			{SQLiteCol: "maxOrderTier3", PostgresCol: "max_order_tier3"},
			{SQLiteCol: "minOrder1", PostgresCol: "min_order1"},
			{SQLiteCol: "maxOrder1", PostgresCol: "max_order1"},
			{SQLiteCol: "createdAt", PostgresCol: "created_at", Transform: transformDateTime},
			{SQLiteCol: "updatedAt", PostgresCol: "updated_at", Transform: transformDateTime},
		},
	}
}

func getWebhookLogMapping() TableMapping {
	return TableMapping{
		SQLiteTable:   "WebhookLog",
		PostgresTable: "webhook_logs",
		HasTenantID:   true,
		Columns: []ColumnMapping{
			{SQLiteCol: "id", PostgresCol: "id"},
			{SQLiteCol: "tenantId", PostgresCol: "tenant_id"},
			{SQLiteCol: "platform", PostgresCol: "platform"},
			{SQLiteCol: "eventType", PostgresCol: "event_type"},
			{SQLiteCol: "payload", PostgresCol: "payload", Transform: transformJSON},
			{SQLiteCol: "headers", PostgresCol: "headers", Transform: transformJSON},
			{SQLiteCol: "status", PostgresCol: "status"},
			{SQLiteCol: "errorMsg", PostgresCol: "error_msg"},
			{SQLiteCol: "processedAt", PostgresCol: "processed_at", Transform: transformDateTime},
			{SQLiteCol: "createdAt", PostgresCol: "created_at", Transform: transformDateTime},
		},
	}
}

func getOAuthStateMapping() TableMapping {
	return TableMapping{
		SQLiteTable:   "OAuthState",
		PostgresTable: "oauth_states",
		HasTenantID:   true,
		Columns: []ColumnMapping{
			{SQLiteCol: "id", PostgresCol: "id"},
			{SQLiteCol: "tenantId", PostgresCol: "tenant_id"},
			{SQLiteCol: "platform", PostgresCol: "platform"},
			{SQLiteCol: "state", PostgresCol: "state"},
			{SQLiteCol: "redirectUrl", PostgresCol: "redirect_url"},
			{SQLiteCol: "metadata", PostgresCol: "metadata", Transform: transformJSON},
			{SQLiteCol: "expiresAt", PostgresCol: "expires_at", Transform: transformDateTime},
			{SQLiteCol: "createdAt", PostgresCol: "created_at", Transform: transformDateTime},
		},
	}
}
