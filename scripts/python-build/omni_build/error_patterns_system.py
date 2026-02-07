"""
System Error Pattern definitions for Omni Build System.

SRP: This module defines system/network/build error patterns.
Service patterns are in error_patterns_services.py.
Infrastructure patterns are in error_patterns.py.
"""
from omni_build.models import ErrorPattern, ErrorSeverity


def get_network_dns_patterns() -> list[ErrorPattern]:
    """Network and DNS error patterns."""
    return [
        ErrorPattern(
            name="DNSError",
            pattern=r"DNS lookup error|DNS.*name does not exist|SERVFAIL|NXDOMAIN|fetch.*error|lookup.*no such host|dial tcp.*lookup",
            description="DNS resolution failure",
            severity=ErrorSeverity.HIGH,
            fix_function="repair_dns",
        ),
        ErrorPattern(
            name="AlpineRepoError",
            pattern=r"fetching https://dl-cdn\.alpinelinux\.org.*temporary error|unable to select packages.*alpine|apk.*error",
            description="Alpine repository network error",
            severity=ErrorSeverity.MEDIUM,
            fix_function="repair_alpine_repo",
        ),
        ErrorPattern(
            name="DockerRegistryError",
            pattern=r"registry-1\.docker\.io.*no such host|failed to do request.*registry|registry.*connection|pull.*manifest.*error",
            description="Docker registry connection error",
            severity=ErrorSeverity.HIGH,
            fix_function="repair_docker_registry",
        ),
        ErrorPattern(
            name="NetworkError",
            pattern=r"network.*error|connect.*refused|ECONNREFUSED|timeout|timed out",
            description="Network connectivity issue",
            severity=ErrorSeverity.MEDIUM,
            fix_function="repair_network",
        ),
    ]


def get_resource_patterns() -> list[ErrorPattern]:
    """System resource error patterns."""
    return [
        ErrorPattern(
            name="DiskSpaceError",
            pattern=r"no space left|disk.*full|ENOSPC",
            description="Disk space exhausted",
            severity=ErrorSeverity.CRITICAL,
            fix_function="cleanup_disk_space",
        ),
        ErrorPattern(
            name="OOMError",
            pattern=r"exit.*code.*137|killed|out of memory|OOM|ENOMEM",
            description="Out of memory",
            severity=ErrorSeverity.CRITICAL,
            fix_function="repair_oom",
        ),
        ErrorPattern(
            name="PortConflictError",
            pattern=r"port.*already.*use|address already in use|EADDRINUSE",
            description="Port conflict",
            severity=ErrorSeverity.MEDIUM,
            fix_function="repair_port_conflict",
        ),
        # === SILENT ERRORS - System masih running tapi ada pressure ===
        ErrorPattern(
            name="HighCPUUsage",
            pattern=r"cpu.*100%|high cpu|cpu.*overload|load average.*high|cpu throttl",
            description="CPU usage tinggi - performance degraded",
            severity=ErrorSeverity.MEDIUM,
            fix_function="log_high_cpu",
        ),
        ErrorPattern(
            name="MemoryPressure",
            pattern=r"memory pressure|low memory|memory.*cgroup|memory.*limit|swap.*active|thrashing",
            description="System memory pressure - potential OOM soon",
            severity=ErrorSeverity.HIGH,
            fix_function="repair_memory_pressure",
        ),
        ErrorPattern(
            name="DiskLatency",
            pattern=r"disk.*latency|io.*wait|slow.*disk|high.*iowait|disk.*bottleneck|storage.*slow",
            description="Disk I/O lambat - storage bottleneck",
            severity=ErrorSeverity.MEDIUM,
            fix_function="log_disk_latency",
        ),
        ErrorPattern(
            name="DiskSpaceWarning",
            pattern=r"disk.*[89][0-9]%|disk usage.*high|low disk space|filesystem.*full",
            description="Disk space warning (80-99%)",
            severity=ErrorSeverity.MEDIUM,
            fix_function="cleanup_disk_space",
        ),
    ]


def get_build_patterns() -> list[ErrorPattern]:
    """Build and NPM error patterns."""
    return [
        ErrorPattern(
            name="NpmIntegrityError",
            pattern=r"integrity checksum|EINTEGRITY|sha512|sha1.*mismatch",
            description="NPM integrity checksum mismatch",
            severity=ErrorSeverity.MEDIUM,
            fix_function="repair_npm_integrity",
        ),
    ]


def get_code_error_patterns() -> list[ErrorPattern]:
    """Code errors that cannot be auto-fixed."""
    return [
        ErrorPattern(
            name="TypeScriptCasingError",
            pattern=r"error TS1149:|error TS1261:|differs from.*only in casing|File name.*differs from already included file name",
            description="TypeScript file casing mismatch (Windows case-insensitivity issue)",
            severity=ErrorSeverity.HIGH,
            fix_function=None,
            is_code_error=True,
        ),
        ErrorPattern(
            name="TypeScriptError",
            pattern=r"error TS\d+:|Cannot find module|has no exported member|is not assignable to|Object is of type 'unknown'",
            description="TypeScript compilation error",
            severity=ErrorSeverity.HIGH,
            fix_function=None,
            is_code_error=True,
        ),
        ErrorPattern(
            name="ImportResolutionError",
            pattern=r"Could not resolve|Failed to resolve import|Module not found.*from",
            description="Import/Module resolution error",
            severity=ErrorSeverity.HIGH,
            fix_function=None,
            is_code_error=True,
        ),
        ErrorPattern(
            name="SyntaxError",
            pattern=r"SyntaxError|Unexpected token|Parse error",
            description="JavaScript/TypeScript syntax error",
            severity=ErrorSeverity.HIGH,
            fix_function=None,
            is_code_error=True,
        ),
        ErrorPattern(
            name="ViteError",
            pattern=r"vite.*error|rollup.*error",
            description="Vite/Rollup build error",
            severity=ErrorSeverity.HIGH,
            fix_function=None,
            is_code_error=True,
        ),
    ]
