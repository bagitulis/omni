"""
Service Error Pattern definitions for Omni Build System.

SRP: This module defines service-specific error patterns (Nginx, Cloudflared, etc).
System patterns are in error_patterns_system.py.
Infrastructure patterns are in error_patterns.py.
"""
from omni_build.models import ErrorPattern, ErrorSeverity


def get_nginx_patterns() -> list[ErrorPattern]:
    """Nginx error patterns."""
    return [
        ErrorPattern(
            name="NginxUpstreamError",
            pattern=r"upstream.*timed out|upstream prematurely closed|no live upstreams|connect\(\) failed.*upstream|upstream.*502|upstream.*504",
            description="Nginx upstream (backend/frontend) not reachable",
            severity=ErrorSeverity.HIGH,
            fix_function="repair_nginx_upstream",
        ),
        ErrorPattern(
            name="NginxConfigError",
            pattern=r"nginx.*emerg|nginx.*error.*conf|unknown directive|nginx: \[emerg\]",
            description="Nginx configuration error",
            severity=ErrorSeverity.CRITICAL,
            fix_function="restart_nginx",
        ),
        ErrorPattern(
            name="NginxBindError",
            pattern=r"nginx.*bind\(\).*failed|nginx.*Address already in use|0\.0\.0\.0:80.*failed",
            description="Nginx cannot bind to port (port conflict)",
            severity=ErrorSeverity.HIGH,
            fix_function="repair_port_conflict",
        ),
        # === SILENT ERRORS ===
        ErrorPattern(
            name="NginxWorkerError",
            pattern=r"worker process.*exited|worker.*abnormally|signal process.*exited|worker.*terminated",
            description="Nginx worker process crash - masih serving tapi tidak stabil",
            severity=ErrorSeverity.HIGH,
            fix_function="restart_nginx",
        ),
        ErrorPattern(
            name="NginxSlowUpstream",
            pattern=r"upstream.*slow|upstream response time.*exceeded|client intended to send too large body|upstream.*buffering",
            description="Nginx upstream response lambat",
            severity=ErrorSeverity.MEDIUM,
            fix_function="log_nginx_slow_upstream",
        ),
        ErrorPattern(
            name="NginxBufferOverflow",
            pattern=r"client request body.*too large|client intended to send too large body|too big.*body|large.*body.*error",
            description="Nginx buffer overflow - request terlalu besar",
            severity=ErrorSeverity.MEDIUM,
            fix_function="log_nginx_buffer_issue",
        ),
        ErrorPattern(
            name="NginxSSLError",
            pattern=r"SSL.*error|SSL_do_handshake.*failed|certificate.*error|SSL.*failed|ssl_stapling.*error",
            description="Nginx SSL/TLS error",
            severity=ErrorSeverity.MEDIUM,
            fix_function="log_nginx_ssl_issue",
        ),
    ]


def get_cloudflared_patterns() -> list[ErrorPattern]:
    """Cloudflare tunnel error patterns."""
    return [
        ErrorPattern(
            name="CloudflaredTunnelError",
            pattern=r"cloudflared.*error|tunnel.*connection failed|Unable to reach.*cloudflare|failed to connect to origin|ERR.*tunnel",
            description="Cloudflare tunnel connection error",
            severity=ErrorSeverity.HIGH,
            fix_function="restart_cloudflared",
        ),
        ErrorPattern(
            name="CloudflaredAuthError",
            pattern=r"cloudflared.*authentication|tunnel.*credentials|invalid.*token|tunnel.*not found|failed to fetch tunnel",
            description="Cloudflare tunnel authentication failed",
            severity=ErrorSeverity.CRITICAL,
            fix_function="restart_cloudflared",
        ),
        ErrorPattern(
            name="CloudflaredDNSError",
            pattern=r"cloudflared.*DNS|failed to lookup.*cloudflare|SERVFAIL.*cloudflare|tunnel.*hostname",
            description="Cloudflare DNS resolution failed",
            severity=ErrorSeverity.MEDIUM,
            fix_function="repair_dns",
        ),
        # === SILENT ERRORS ===
        ErrorPattern(
            name="CloudflaredReconnection",
            pattern=r"reconnecting.*tunnel|tunnel.*reconnect|connection.*retry|retrying.*connection|lost connection.*tunnel",
            description="Cloudflare tunnel reconnecting - temporary instability",
            severity=ErrorSeverity.MEDIUM,
            fix_function="log_cloudflared_reconnection",
        ),
        ErrorPattern(
            name="CloudflaredQuotaExceeded",
            pattern=r"quota.*exceeded|rate.*limit|too many requests|tunnel.*limit|concurrent.*limit",
            description="Cloudflare quota/rate limit exceeded",
            severity=ErrorSeverity.HIGH,
            fix_function="log_cloudflared_quota",
        ),
        ErrorPattern(
            name="CloudflaredOriginError",
            pattern=r"origin.*error|origin.*unreachable|failed to connect to origin|unable to reach origin|origin.*timeout",
            description="Cloudflare cannot reach origin (backend/frontend)",
            severity=ErrorSeverity.HIGH,
            fix_function="repair_cloudflared_origin",
        ),
    ]


def get_pgbouncer_patterns() -> list[ErrorPattern]:
    """PgBouncer connection pooler error patterns."""
    return [
        ErrorPattern(
            name="PgBouncerConnectionError",
            pattern=r"pgbouncer.*connection.*refused|pgbouncer.*failed.*connect|cannot connect to pgbouncer|6432.*refused",
            description="PgBouncer connection pool not reachable",
            severity=ErrorSeverity.HIGH,
            fix_function="restart_pgbouncer",
        ),
        ErrorPattern(
            name="PgBouncerPoolError",
            pattern=r"pgbouncer.*pool|no more connections allowed|too many clients|connection pool exhausted",
            description="PgBouncer connection pool exhausted",
            severity=ErrorSeverity.MEDIUM,
            fix_function="restart_pgbouncer",
        ),
        ErrorPattern(
            name="PgBouncerAuthError",
            pattern=r"pgbouncer.*auth.*failed|pgbouncer.*password|pgbouncer.*FATAL|password authentication failed.*pgbouncer",
            description="PgBouncer authentication error",
            severity=ErrorSeverity.HIGH,
            fix_function="restart_pgbouncer",
        ),
        # === SILENT ERRORS ===
        ErrorPattern(
            name="PgBouncerPoolSaturation",
            pattern=r"pool.*full|waiting.*connection|client.*queue|server connection.*limit|reserve_pool_size",
            description="PgBouncer connection pool hampir penuh",
            severity=ErrorSeverity.HIGH,
            fix_function="repair_pgbouncer_pool",
        ),
        ErrorPattern(
            name="PgBouncerQueryTimeout",
            pattern=r"query.*timeout|query_wait_timeout|server.*check.*failed|login.*timed out",
            description="PgBouncer query timeout",
            severity=ErrorSeverity.MEDIUM,
            fix_function="log_pgbouncer_timeout",
        ),
        ErrorPattern(
            name="PgBouncerServerDisconnect",
            pattern=r"server.*disconnect|closing.*because.*server|unexpected eof|pooler error",
            description="PgBouncer server disconnection - potential PostgreSQL issue",
            severity=ErrorSeverity.HIGH,
            fix_function="repair_pgbouncer_disconnect",
        ),
    ]


def get_pgweb_patterns() -> list[ErrorPattern]:
    """PgWeb database viewer error patterns."""
    return [
        ErrorPattern(
            name="PgWebConnectionError",
            pattern=r"pgweb.*connection|pgweb.*error|pgweb.*failed|Cannot connect to database.*pgweb",
            description="PgWeb database viewer connection error",
            severity=ErrorSeverity.LOW,
            fix_function="restart_pgweb",
        ),
        ErrorPattern(
            name="PgWebPortError",
            pattern=r"pgweb.*8081.*in use|pgweb.*address already|pgweb.*bind.*failed",
            description="PgWeb port conflict",
            severity=ErrorSeverity.LOW,
            fix_function="restart_pgweb",
        ),
    ]


def get_backend_patterns() -> list[ErrorPattern]:
    """Backend (Go) error patterns."""
    return [
        ErrorPattern(
            name="BackendHealthCheckTimeout",
            pattern=r"backend.*did not become healthy|backend.*Timeout after|Health checks failed|services not responding",
            description="Backend service health check timeout (Go backend initializing)",
            severity=ErrorSeverity.MEDIUM,
            fix_function="wait_for_backend_init",
        ),
        # === SILENT ERRORS ===
        ErrorPattern(
            name="BackendGoroutineLeak",
            pattern=r"goroutine.*leak|too many goroutines|goroutine count.*exceeded|runtime.*goroutine.*growing",
            description="Go goroutine leak - memory akan terus naik",
            severity=ErrorSeverity.HIGH,
            fix_function="restart_backend",
        ),
        ErrorPattern(
            name="BackendDBPoolExhausted",
            pattern=r"connection pool exhausted|all pool connections.*busy|acquiring conn.*timeout|sql.*too many connections|gorm.*connection",
            description="Backend database connection pool habis",
            severity=ErrorSeverity.HIGH,
            fix_function="repair_backend_db_pool",
        ),
        ErrorPattern(
            name="BackendSlowQuery",
            pattern=r"slow query.*ms|query took.*seconds|slow sql|SLOW.*SQL|execution time exceeded",
            description="Backend slow query detected",
            severity=ErrorSeverity.MEDIUM,
            fix_function="log_backend_slow_query",
        ),
        ErrorPattern(
            name="BackendPanicRecover",
            pattern=r"panic.*recover|runtime error|nil pointer|index out of range|slice bounds out of range",
            description="Go panic (recovered) - potential bug",
            severity=ErrorSeverity.HIGH,
            fix_function="log_backend_panic",
        ),
    ]
