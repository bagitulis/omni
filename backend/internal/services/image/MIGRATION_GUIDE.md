# Image Management System Migration Guide

## Overview

This guide documents the new unified image management system with content-based deduplication and multi-size thumbnail generation.

## Database Schema Changes

### Updated Image Model

```go
type Image struct {
    ID          int64     `gorm:"primaryKey" json:"id"`
    TenantID    string    `gorm:"index;not null" json:"tenant_id"`
    ContentHash string    `gorm:"size:64;index;not null" json:"content_hash"`
    OriginalURL string    `gorm:"index" json:"original_url,omitempty"`
    LocalPath   string    `gorm:"not null" json:"local_path"` // base path without size suffix
    Width       int       `json:"width"`
    Height      int       `json:"height"`
    FileSize    int64     `json:"file_size"`
    RefCount    int       `gorm:"default:1" json:"ref_count"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### Removed Fields

- `Filename` - not needed with content-hash based storage
- `Size` - renamed to `FileSize` for clarity
- `MimeType` - all stored as WebP, not needed in DB
- `Category` - removed, using directory structure instead
- `ProductID` - moved to join table
- `Thumbnails` - paths are computed dynamically
- `DeletedAt` - hard delete when ref_count <= 0

### Required Migration

```sql
-- Create unique constraint for deduplication
CREATE UNIQUE INDEX idx_images_tenant_hash ON images(tenant_id, content_hash);

-- For existing data, you may need to:
-- 1. Calculate content_hash for existing images
-- 2. Migrate file paths to new structure
-- 3. Update ref_count based on actual usage
```

## New File Structure

```
uploads/
└── {tenant_id}/
    └── images/
        └── {content_hash}/
            ├── original.webp  (max 1200x1200, quality 85)
            ├── medium.webp    (max 300x300, quality 80)
            └── thumb.webp     (max 80x80, quality 75)
```

## API Usage

### Creating Image Manager

```go
import "github.com/omni/backend/internal/services/image"

mgr := image.NewManager(db, "uploads")
```

### Caching Remote Image

```go
img, err := mgr.CacheImage(ctx, tenantID, "https://example.com/image.jpg")
if err != nil {
    // Handle error
}

// Get all three thumbnail paths
paths := mgr.GetPaths(img)
fmt.Println(paths.Thumb)    // /uploads/tenant/images/hash/thumb.webp
fmt.Println(paths.Medium)   // /uploads/tenant/images/hash/medium.webp
fmt.Println(paths.Original) // /uploads/tenant/images/hash/original.webp
```

### Reference Counting

```go
// Increment when linking to a product/entity
err := mgr.AddRef(ctx, imageID)

// Decrement when unlinking
err := mgr.RemoveRef(ctx, imageID)

// Cleanup orphaned images (ref_count <= 0)
deleted, err := mgr.CleanupOrphans(ctx, tenantID)
fmt.Printf("Deleted %d orphaned images\n", deleted)
```

## Deduplication Logic

1. **URL Check**: First checks if image with same `original_url` exists
2. **Download**: Downloads image from remote URL (max 10MB, 30s timeout)
3. **Hash Check**: Calculates SHA256 hash of raw bytes, checks if image with same `content_hash` exists
4. **Storage**: If new, creates directory and saves 3 thumbnail sizes
5. **Database**: Creates record with `ref_count=1` or increments existing

## Benefits

1. **Space Efficiency**: Identical images stored only once per tenant
2. **Performance**: Multiple thumbnail sizes pre-generated
3. **Consistency**: All images in WebP format with controlled quality
4. **Safety**: Reference counting prevents accidental deletion
5. **Scalability**: Content-hash based paths distribute files evenly

## Migration Notes

### Compatibility with Existing Services

- `ProductImageCacheService` - Still works, not modified
- `DedupService` - Updated to match new schema
- `ReferenceService` - Updated signature (uint → int64)
- `ImageHandler` - Updated to match new schema

### Type Changes

- Image ID changed from `uint` to `int64` for consistency
- Reference counting now uses `int64` throughout

## Example: Product Image Flow

```go
// When cloning a product with images
var processedImages []models.Image

for _, imageURL := range productImageURLs {
    img, err := mgr.CacheImage(ctx, tenantID, imageURL)
    if err != nil {
        log.Error().Err(err).Str("url", imageURL).Msg("Failed to cache image")
        continue
    }
    processedImages = append(processedImages, *img)
}

// Create join table entries (ref_count already incremented by CacheImage)
for i, img := range processedImages {
    link := MasterProductImage{
        ProductID: productID,
        ImageID:   uint(img.ID),
        SortOrder: i,
        Role:      "gallery",
    }
    db.Create(&link)
}
```

## Cleanup Strategy

```go
// Periodic cleanup (run via cron)
func CleanupOrphanedImages(ctx context.Context, mgr image.Manager, tenantID string) {
    deleted, err := mgr.CleanupOrphans(ctx, tenantID)
    if err != nil {
        log.Error().Err(err).Msg("Cleanup failed")
        return
    }
    log.Info().Int("deleted", deleted).Msg("Cleanup completed")
}
```

## Performance Considerations

1. **Concurrent Safety**: CacheImage uses mutex to prevent race conditions during hash check
2. **Transaction Safety**: All database operations use transactions where needed
3. **File System**: Hash-based directory structure distributes files evenly
4. **Index Usage**: Composite index on (tenant_id, content_hash) for fast lookups

## Future Enhancements

- [ ] Background job for orphan cleanup
- [ ] CDN integration for public URLs
- [ ] Image optimization (smart cropping, lazy loading)
- [ ] Metadata extraction (EXIF, color palette)
- [ ] Multi-tenant sharing (optional cross-tenant dedup)
