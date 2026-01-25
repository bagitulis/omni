package shopeesdk

import "context"

// GetWalletTransactions fetches wallet transaction history.
func (c *Client) GetWalletTransactions(ctx context.Context, req GetWalletTransactionRequest) (*WalletTransactionResponse, error) {
	if err := apiReady(ctx); err != nil {
		return nil, err
	}
	return c.api.GetWalletTransactions(req)
}
