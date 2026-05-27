package analytics

import (
	"strconv"
	"time"

	tiktokPkg "github.com/omni/backend/pkg/tiktok"
)

type TiktokEscrowClient interface {
	SearchOrders(req tiktokPkg.OrderSearchRequest, pageSize int, pageToken string) (*tiktokPkg.OrderSearchResponse, error)
	GetOrderDetail(orderIDs []string) (*tiktokPkg.OrderDetailResponse, error)
	GetOrderTransactions(orderID string) (*tiktokPkg.OrderTransactionResponse, error)
	GetOrderTransactionsV202309(orderID string) (*tiktokPkg.OrderTransactionResponse, error)
}

type TiktokEscrowTables struct {
	OrderTable string
	ItemTable  string
	SyncTable  string
}

func tiktokEscrowTables() TiktokEscrowTables {
	return TiktokEscrowTables{
		OrderTable: "tiktok_escrow_orders",
		ItemTable:  "tiktok_escrow_items",
		SyncTable:  "tiktok_escrow_sync",
	}
}

func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

func parseTiktokTimestamp(ts int64) time.Time {
	if ts <= 0 {
		return time.Time{}
	}
	return time.Unix(ts, 0).UTC()
}

func normalizeShippingFee(fee float64) float64 {
	if fee < 0 {
		return -fee
	}
	return fee
}
