"""
TikTok Ads Tables Creation Script
Add TikTok Ads tables to existing database WITHOUT resetting data

Per AGENTS.MD: Use Python for non-destructive schema changes
"""

import sqlite3
import os
from pathlib import Path

# Database paths
BACKEND_DIR = Path(__file__).parent.parent
DATABASES_DIR = BACKEND_DIR / "config" / "databases"

TENANT_DATABASES = [
    "yumna_bertigamart.db",
    "tika_nusseyba.db",
]

# TikTok Ads table creation SQL
TIKTOK_ADS_TABLES = [
    # TiktokAdsUploadBatch
    """
    CREATE TABLE IF NOT EXISTS TiktokAdsUploadBatch (
        id TEXT PRIMARY KEY,
        tenantId TEXT NOT NULL,
        fileName TEXT NOT NULL,
        periodStart DATETIME NOT NULL,
        periodEnd DATETIME NOT NULL,
        totalRows INTEGER NOT NULL DEFAULT 0,
        insertedRows INTEGER DEFAULT 0,
        skippedRows INTEGER DEFAULT 0,
        updatedRows INTEGER DEFAULT 0,
        status TEXT NOT NULL DEFAULT 'pending',
        errorMessage TEXT,
        uploadedBy TEXT,
        createdAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updatedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    )
    """,
    "CREATE INDEX IF NOT EXISTS idx_tiktokads_batch_tenant ON TiktokAdsUploadBatch(tenantId)",
    "CREATE INDEX IF NOT EXISTS idx_tiktokads_batch_period ON TiktokAdsUploadBatch(periodStart, periodEnd)",
    
    # TiktokAdsCreativeData
    """
    CREATE TABLE IF NOT EXISTS TiktokAdsCreativeData (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        tenantId TEXT NOT NULL,
        uploadBatchId TEXT,
        periodStart DATETIME NOT NULL,
        periodEnd DATETIME NOT NULL,
        campaignId TEXT NOT NULL,
        campaignName TEXT NOT NULL,
        productId TEXT NOT NULL,
        creativeType TEXT NOT NULL,
        videoTitle TEXT,
        videoId TEXT,
        tiktokAccount TEXT,
        postingTime DATETIME,
        status TEXT,
        authorizationType TEXT,
        cost REAL NOT NULL DEFAULT 0,
        ordersSku INTEGER NOT NULL DEFAULT 0,
        costPerOrder REAL NOT NULL DEFAULT 0,
        grossRevenue REAL NOT NULL DEFAULT 0,
        roi REAL NOT NULL DEFAULT 0,
        impressions INTEGER NOT NULL DEFAULT 0,
        clicks INTEGER NOT NULL DEFAULT 0,
        ctr REAL NOT NULL DEFAULT 0,
        conversionRate REAL NOT NULL DEFAULT 0,
        watchRate2s REAL,
        watchRate6s REAL,
        watchRate25pct REAL,
        watchRate50pct REAL,
        watchRate75pct REAL,
        watchRate100pct REAL,
        currency TEXT NOT NULL DEFAULT 'IDR',
        createdAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updatedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (uploadBatchId) REFERENCES TiktokAdsUploadBatch(id) ON DELETE SET NULL
    )
    """,
    "CREATE UNIQUE INDEX IF NOT EXISTS idx_tiktokads_unique_creative ON TiktokAdsCreativeData(tenantId, campaignId, productId, videoId, periodStart, periodEnd)",
    "CREATE INDEX IF NOT EXISTS idx_tiktokads_creative_tenant ON TiktokAdsCreativeData(tenantId)",
    "CREATE INDEX IF NOT EXISTS idx_tiktokads_creative_campaign ON TiktokAdsCreativeData(campaignId)",
    "CREATE INDEX IF NOT EXISTS idx_tiktokads_creative_product ON TiktokAdsCreativeData(productId)",
    "CREATE INDEX IF NOT EXISTS idx_tiktokads_creative_type ON TiktokAdsCreativeData(creativeType)",
    "CREATE INDEX IF NOT EXISTS idx_tiktokads_creative_period ON TiktokAdsCreativeData(periodStart, periodEnd)",
    "CREATE INDEX IF NOT EXISTS idx_tiktokads_creative_batch ON TiktokAdsCreativeData(uploadBatchId)",
    
    # TiktokAdsProductSummary
    """
    CREATE TABLE IF NOT EXISTS TiktokAdsProductSummary (
        id TEXT PRIMARY KEY,
        tenantId TEXT NOT NULL,
        productId TEXT NOT NULL,
        periodType TEXT NOT NULL,
        periodDate DATETIME NOT NULL,
        totalCost REAL NOT NULL DEFAULT 0,
        totalOrders INTEGER NOT NULL DEFAULT 0,
        totalRevenue REAL NOT NULL DEFAULT 0,
        avgRoi REAL NOT NULL DEFAULT 0,
        avgCtr REAL NOT NULL DEFAULT 0,
        avgConversionRate REAL NOT NULL DEFAULT 0,
        totalImpressions INTEGER NOT NULL DEFAULT 0,
        totalClicks INTEGER NOT NULL DEFAULT 0,
        bestCreativeType TEXT,
        topVideoId TEXT,
        topVideoTitle TEXT,
        roiTrend TEXT,
        costTrend TEXT,
        ordersTrend TEXT,
        createdAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updatedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    )
    """,
    "CREATE UNIQUE INDEX IF NOT EXISTS idx_tiktokads_summary_unique ON TiktokAdsProductSummary(tenantId, productId, periodType, periodDate)",
    "CREATE INDEX IF NOT EXISTS idx_tiktokads_summary_tenant ON TiktokAdsProductSummary(tenantId)",
    "CREATE INDEX IF NOT EXISTS idx_tiktokads_summary_product ON TiktokAdsProductSummary(productId)",
    "CREATE INDEX IF NOT EXISTS idx_tiktokads_summary_period ON TiktokAdsProductSummary(periodType, periodDate)",
    
    # TiktokAdsMLPrediction
    """
    CREATE TABLE IF NOT EXISTS TiktokAdsMLPrediction (
        id TEXT PRIMARY KEY,
        tenantId TEXT NOT NULL,
        productId TEXT NOT NULL,
        predictionDate DATETIME NOT NULL,
        targetMonth INTEGER NOT NULL,
        targetYear INTEGER NOT NULL,
        predictedRoi REAL,
        roiConfidence REAL,
        recommendedBudget REAL,
        budgetReason TEXT,
        performanceScore TEXT,
        creativeRecommendation TEXT,
        riskLevel TEXT,
        riskFactors TEXT,
        modelVersion TEXT,
        createdAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updatedAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
    )
    """,
    "CREATE UNIQUE INDEX IF NOT EXISTS idx_tiktokads_ml_unique ON TiktokAdsMLPrediction(tenantId, productId, targetMonth, targetYear)",
    "CREATE INDEX IF NOT EXISTS idx_tiktokads_ml_tenant ON TiktokAdsMLPrediction(tenantId)",
    "CREATE INDEX IF NOT EXISTS idx_tiktokads_ml_product ON TiktokAdsMLPrediction(productId)",
    "CREATE INDEX IF NOT EXISTS idx_tiktokads_ml_target ON TiktokAdsMLPrediction(targetMonth, targetYear)",
]


def table_exists(cursor, table_name):
    """Check if table already exists"""
    cursor.execute(
        "SELECT name FROM sqlite_master WHERE type='table' AND name=?",
        (table_name,)
    )
    return cursor.fetchone() is not None


def add_tiktok_ads_tables(db_path):
    """Add TikTok Ads tables to database without affecting existing data"""
    db_name = os.path.basename(db_path)
    print(f"\n📁 Processing: {db_name}")
    
    if not os.path.exists(db_path):
        print(f"   ❌ Database not found: {db_path}")
        return False
    
    conn = sqlite3.connect(db_path)
    cursor = conn.cursor()
    
    try:
        # Check existing tables
        existing = []
        for table in ["TiktokAdsUploadBatch", "TiktokAdsCreativeData", 
                      "TiktokAdsProductSummary", "TiktokAdsMLPrediction"]:
            if table_exists(cursor, table):
                existing.append(table)
        
        if existing:
            print(f"   ℹ️  Tables already exist: {', '.join(existing)}")
        
        # Create tables
        created = 0
        for sql in TIKTOK_ADS_TABLES:
            try:
                cursor.execute(sql)
                if "CREATE TABLE" in sql:
                    created += 1
            except sqlite3.Error as e:
                # Ignore "already exists" errors
                if "already exists" not in str(e):
                    print(f"   ⚠️  SQL error: {e}")
        
        conn.commit()
        print(f"   ✅ TikTok Ads tables ensured (executed {len(TIKTOK_ADS_TABLES)} statements)")
        
        # Verify
        cursor.execute("SELECT name FROM sqlite_master WHERE type='table' AND name LIKE 'TiktokAds%'")
        tiktok_tables = [row[0] for row in cursor.fetchall()]
        print(f"   📊 TikTok Ads tables present: {', '.join(tiktok_tables)}")
        
        return True
        
    except Exception as e:
        print(f"   ❌ Error: {e}")
        conn.rollback()
        return False
    finally:
        conn.close()


def main():
    print("=" * 50)
    print("TikTok Ads Tables Creation (Non-Destructive)")
    print("=" * 50)
    
    success = 0
    failed = 0
    
    for db_name in TENANT_DATABASES:
        db_path = DATABASES_DIR / db_name
        if add_tiktok_ads_tables(str(db_path)):
            success += 1
        else:
            failed += 1
    
    print("\n" + "=" * 50)
    print(f"Complete: {success} succeeded, {failed} failed")
    print("=" * 50)


if __name__ == "__main__":
    main()
