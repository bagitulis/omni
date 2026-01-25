package shopeesdk

import (
	"context"
	"errors"
)

// GetProducts returns item IDs for a shop.
func (c *Client) GetProducts(ctx context.Context, offset, limit int) (*ProductListResponse, error) {
	if err := apiReady(ctx); err != nil {
		return nil, err
	}
	return c.api.GetProductList(offset, limit)
}

// GetProductBaseInfo returns base info for up to 50 item IDs.
func (c *Client) GetProductBaseInfo(ctx context.Context, itemIDs []int64) (*ProductDetailResponse, error) {
	if err := apiReady(ctx); err != nil {
		return nil, err
	}
	if len(itemIDs) == 0 {
		return nil, errors.New("item_ids is required")
	}
	return c.api.GetProductDetail(itemIDs)
}

// GetProductInfoWithImages fetches base info plus image URLs.
func (c *Client) GetProductInfoWithImages(ctx context.Context, itemIDs []int64) (*ProductDetailWithImagesResponse, error) {
	if err := apiReady(ctx); err != nil {
		return nil, err
	}
	if len(itemIDs) == 0 {
		return nil, errors.New("item_ids is required")
	}
	return c.api.GetProductDetailWithImages(itemIDs)
}

// GetModelList returns models/SKUs for an item.
func (c *Client) GetModelList(ctx context.Context, itemID int64) (*ModelListResponse, error) {
	if err := apiReady(ctx); err != nil {
		return nil, err
	}
	if itemID == 0 {
		return nil, errors.New("item_id is required")
	}
	return c.api.GetModelList(itemID)
}

// CreateProduct creates a new item.
func (c *Client) CreateProduct(ctx context.Context, req CreateProductRequest) (*CreateProductResponse, error) {
	if err := apiReady(ctx); err != nil {
		return nil, err
	}
	return c.api.CreateProduct(req)
}

// UpdateProduct updates an item.
func (c *Client) UpdateProduct(ctx context.Context, req UpdateProductRequest) (*UpdateProductResponse, error) {
	if err := apiReady(ctx); err != nil {
		return nil, err
	}
	if req.ItemID == 0 {
		return nil, errors.New("item_id is required")
	}
	return c.api.UpdateProduct(req)
}

// DeleteProduct deletes an item.
func (c *Client) DeleteProduct(ctx context.Context, itemID int64) (*DeleteProductResponse, error) {
	if err := apiReady(ctx); err != nil {
		return nil, err
	}
	if itemID == 0 {
		return nil, errors.New("item_id is required")
	}
	return c.api.DeleteProduct(itemID)
}

// UpdateStock updates stock for models inside an item.
func (c *Client) UpdateStock(ctx context.Context, req UpdateStockRequest) (*UpdateStockResponse, error) {
	if err := apiReady(ctx); err != nil {
		return nil, err
	}
	if req.ItemID == 0 {
		return nil, errors.New("item_id is required")
	}
	return c.api.UpdateStock(req)
}

// UpdatePrice updates price for models inside an item.
func (c *Client) UpdatePrice(ctx context.Context, req UpdatePriceRequest) (*UpdatePriceResponse, error) {
	if err := apiReady(ctx); err != nil {
		return nil, err
	}
	if req.ItemID == 0 {
		return nil, errors.New("item_id is required")
	}
	return c.api.UpdatePrice(req)
}

// UploadImage uploads an image and returns IDs usable in create/update calls.
func (c *Client) UploadImage(ctx context.Context, imageBytes []byte) (*UploadImageResponse, error) {
	if err := apiReady(ctx); err != nil {
		return nil, err
	}
	if len(imageBytes) == 0 {
		return nil, errors.New("image bytes are required")
	}
	return c.api.UploadImage(imageBytes)
}
