package handlers

import "gorm.io/gorm"

// lazadaOrderInfo represents Lazada order info needed for bulk ship.
type lazadaOrderInfo struct {
	ShippingCarrier string `gorm:"column:shipping_carrier"`
	OrderStatus     string `gorm:"column:order_status"`
}

// fetchLazadaItemIDsFromDB fetches item IDs for a Lazada order.
func fetchLazadaItemIDsFromDB(db *gorm.DB, orderSN string) ([]string, error) {
	var items []struct {
		ItemID string `gorm:"column:item_id"`
	}
	if err := db.Table("LazadaOrderItem").Select("item_id").Where("order_sn = ?", orderSN).Scan(&items).Error; err != nil {
		return nil, err
	}

	ids := make([]string, len(items))
	for i, item := range items {
		ids[i] = item.ItemID
	}
	return ids, nil
}

// fetchLazadaOrderInfoFromDB fetches order info for a Lazada order.
func fetchLazadaOrderInfoFromDB(db *gorm.DB, orderSN string) (*lazadaOrderInfo, error) {
	var info lazadaOrderInfo
	if err := db.Table("LazadaOrder").Select("shipping_carrier, order_status").Where("order_sn = ?", orderSN).First(&info).Error; err != nil {
		return nil, err
	}
	return &info, nil
}
