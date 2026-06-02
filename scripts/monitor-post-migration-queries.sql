-- =============================================================================
-- OMNI Post-Migration Monitoring Queries (48h after cutover)
-- Usage: docker exec omni-postgres psql -U omni -d omni_main -f monitor-post-migration-queries.sql
-- =============================================================================

-- 1. Token refresh failure count per platform (from credential_audit_events)
SELECT 
    'TOKEN REFRESH FAILURES' as metric,
    platform,
    COUNT(*) as failure_count,
    MAX(created_at) as last_failure
FROM credential_audit_events 
WHERE event_type = 'token_refresh'
  AND status = 'failed'
  AND created_at >= NOW() - INTERVAL '48 hours'
GROUP BY platform
ORDER BY failure_count DESC;

-- 2. Credential load failure count (connections with invalid status)
SELECT 
    'CREDENTIAL LOAD ISSUES' as metric,
    platform,
    COUNT(*) as issue_count,
    COUNT(CASE WHEN status = 'error' THEN 1 END) as error_status,
    COUNT(CASE WHEN access_token IS NULL OR access_token = '' THEN 1 END) as missing_tokens,
    COUNT(CASE WHEN token_expiry < EXTRACT(EPOCH FROM NOW()) * 1000 THEN 1 END) as expired_tokens
FROM credential_connections
WHERE tenant_id != 'system'
GROUP BY platform
HAVING COUNT(*) > 0;

-- 3. OAuth callback failure count (from audit events)
SELECT 
    'OAUTH CALLBACK FAILURES' as metric,
    platform,
    COUNT(*) as failure_count,
    COUNT(CASE WHEN event_type = 'oauth_callback' AND status = 'failed' THEN 1 END) as callback_failures,
    COUNT(CASE WHEN event_type = 'token_exchange' AND status = 'failed' THEN 1 END) as exchange_failures
FROM credential_audit_events
WHERE created_at >= NOW() - INTERVAL '48 hours'
  AND (event_type LIKE '%callback%' OR event_type LIKE '%exchange%')
  AND status = 'failed'
GROUP BY platform
ORDER BY failure_count DESC;

-- 4. Platform API auth error count (401/403 from audit events)
SELECT 
    'PLATFORM AUTH ERRORS' as metric,
    platform,
    COUNT(*) as error_count,
    COUNT(CASE WHEN metadata->>'http_status' = '401' THEN 1 END) as unauthorized_count,
    COUNT(CASE WHEN metadata->>'http_status' = '403' THEN 1 END) as forbidden_count
FROM credential_audit_events
WHERE created_at >= NOW() - INTERVAL '48 hours'
  AND (metadata->>'http_status' IN ('401', '403') 
       OR event_type LIKE '%auth_error%'
       OR status = 'auth_failed')
GROUP BY platform
ORDER BY error_count DESC;

-- 5. Token expiry monitoring (tokens expiring in next 24h)
SELECT 
    'TOKENS EXPIRING SOON' as metric,
    platform,
    COUNT(*) as expiring_count,
    MIN(TO_TIMESTAMP(token_expiry / 1000)) as earliest_expiry,
    MAX(TO_TIMESTAMP(token_expiry / 1000)) as latest_expiry
FROM credential_connections
WHERE tenant_id != 'system'
  AND status = 'connected'
  AND token_expiry > 0
  AND TO_TIMESTAMP(token_expiry / 1000) BETWEEN NOW() AND NOW() + INTERVAL '24 hours'
GROUP BY platform
ORDER BY expiring_count DESC;

-- 6. Refresh token health (refresh tokens expiring or expired)
SELECT 
    'REFRESH TOKEN HEALTH' as metric,
    platform,
    COUNT(*) as total_connections,
    COUNT(CASE WHEN refresh_expiry > 0 AND TO_TIMESTAMP(refresh_expiry / 1000) < NOW() THEN 1 END) as expired_refresh,
    COUNT(CASE WHEN refresh_expiry > 0 AND TO_TIMESTAMP(refresh_expiry / 1000) BETWEEN NOW() AND NOW() + INTERVAL '7 days' THEN 1 END) as expiring_soon_refresh,
    COUNT(CASE WHEN refresh_expiry = 0 OR refresh_expiry IS NULL THEN 1 END) as missing_refresh_expiry
FROM credential_connections
WHERE tenant_id != 'system'
  AND status = 'connected'
GROUP BY platform;

-- 7. Connection status summary
SELECT 
    'CONNECTION STATUS' as metric,
    platform,
    COUNT(*) as total,
    COUNT(CASE WHEN status = 'connected' THEN 1 END) as connected,
    COUNT(CASE WHEN status = 'disconnected' THEN 1 END) as disconnected,
    COUNT(CASE WHEN status = 'error' THEN 1 END) as error,
    COUNT(CASE WHEN disabled_at IS NOT NULL THEN 1 END) as disabled
FROM credential_connections
WHERE tenant_id != 'system'
GROUP BY platform
ORDER BY platform;

-- 8. Recent audit events (last 24h activity)
SELECT 
    'RECENT ACTIVITY' as metric,
    event_type,
    platform,
    status,
    COUNT(*) as event_count,
    MAX(created_at) as last_event
FROM credential_audit_events
WHERE created_at >= NOW() - INTERVAL '24 hours'
GROUP BY event_type, platform, status
ORDER BY last_event DESC
LIMIT 20;

-- 9. Alert: Any non-zero token refresh failures (CRITICAL)
DO $$
DECLARE
    failure_count INTEGER;
    platform_name TEXT;
BEGIN
    FOR platform_name IN SELECT DISTINCT platform FROM credential_audit_events 
        WHERE event_type = 'token_refresh' AND status = 'failed'
            AND created_at >= NOW() - INTERVAL '48 hours'
    LOOP
        SELECT COUNT(*) INTO failure_count
        FROM credential_audit_events
        WHERE event_type = 'token_refresh'
          AND status = 'failed'
          AND platform = platform_name
          AND created_at >= NOW() - INTERVAL '48 hours';
        
        IF failure_count > 0 THEN
            RAISE NOTICE '🚨 ALERT: % token refresh failures for % in last 48h', failure_count, platform_name;
        END IF;
    END LOOP;
    
    -- Check if no failures found
    IF NOT EXISTS (
        SELECT 1 FROM credential_audit_events
        WHERE event_type = 'token_refresh' AND status = 'failed'
            AND created_at >= NOW() - INTERVAL '48 hours'
    ) THEN
        RAISE NOTICE '✅ No token refresh failures in last 48h';
    END IF;
END $$;

-- 10. Migration health check: canonical vs legacy storage
SELECT 
    'MIGRATION STATUS' as metric,
    'credential_connections' as storage,
    COUNT(*) as total_records,
    COUNT(CASE WHEN access_token IS NOT NULL AND access_token != '' THEN 1 END) as with_tokens,
    COUNT(CASE WHEN created_at >= NOW() - INTERVAL '48 hours' THEN 1 END) as recent_migrations
FROM credential_connections
WHERE tenant_id != 'system'
UNION ALL
SELECT 
    'MIGRATION STATUS' as metric,
    'platform_configs' as storage,
    COUNT(*) as total_records,
    COUNT(CASE WHEN key = 'accessToken' AND value IS NOT NULL AND value != '' THEN 1 END) as with_tokens,
    COUNT(CASE WHEN updated_at >= NOW() - INTERVAL '48 hours' THEN 1 END) as recent_updates
FROM tenant_platform_configs
WHERE key IN ('accessToken', 'refreshToken');
