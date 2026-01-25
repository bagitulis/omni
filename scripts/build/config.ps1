# ============================================
# Build System Configuration
# ============================================

# Project paths
$script:ProjectRoot = (Resolve-Path "$PSScriptRoot\..\..").Path
$script:BackendPath = Join-Path $ProjectRoot "backend"
$script:FrontendPath = Join-Path $ProjectRoot "frontend"
$script:NginxPath = Join-Path $ProjectRoot "nginx"
$script:LogsPath = Join-Path $ProjectRoot "logs"
$script:ScriptsPath = Join-Path $ProjectRoot "scripts\build"

# Docker compose files
$script:DockerComposeTunnel = "docker-compose.tunnel.yml"
$script:DockerComposeLowSpec = "docker-compose.tunnel.lowspec.yml"
$script:DockerComposeStandard = "docker-compose.tunnel.standard.yml"
$script:DockerComposeHighSpec = "docker-compose.tunnel.highspec.yml"

# Container names
$script:ContainerBackend = "omni-backend"
$script:ContainerFrontend = "omni-frontend"
$script:ContainerNginx = "omni-nginx"
$script:ContainerRedis = "omni-redis"

# Timeouts (seconds) - increased for better reliability
$script:DockerStartTimeout = 120  # Increased for pipe stabilization
$script:WslRestartTimeout = 30    # Increased for WSL restart
$script:ContainerHealthTimeout = 60
$script:NpmInstallTimeout = 300
$script:DockerBuildTimeout = 900  # Increased for complex builds

# Retry settings (6 levels of progressive recovery for severe errors)
$script:MaxRetries = 6
$script:RetryDelayBase = 5  # Base delay in seconds (exponential backoff)

# Ports to check and auto-free if occupied
$script:RequiredPorts = @(80, 443, 3000, 5173, 5678)

# Minimum memory (MB) before triggering cleanup
$script:MinMemoryMB = 1024

# Disk space minimum (GB)
$script:MinDiskSpaceGB = 5

# Docker CLI path
$script:DockerCliPath = "C:\Program Files\Docker\Docker\DockerCli.exe"

# Build output mode: "condensed" (default), "full", "quiet"
$script:BuildOutputMode = "condensed"

# RAM specs configuration
$script:RamSpecs = @{
    "lowspec" = @{ Name = "Low Spec"; RAM = "2GB"; ComposeFile = $DockerComposeLowSpec }
    "standard" = @{ Name = "Standard"; RAM = "4GB"; ComposeFile = $DockerComposeStandard }
    "highspec" = @{ Name = "High Spec"; RAM = "8GB+"; ComposeFile = $DockerComposeHighSpec }
}

# Required directories (auto-create before build)
$script:RequiredDirs = @(
    "backend\logs"
    "backend\config\databases"
    "backend\config\static"
    "logs"
)

# Export configuration
function Get-BuildConfig {
    return @{
        ProjectRoot = $script:ProjectRoot
        BackendPath = $script:BackendPath
        FrontendPath = $script:FrontendPath
        NginxPath = $script:NginxPath
        DockerComposeTunnel = $script:DockerComposeTunnel
        MaxRetries = $script:MaxRetries
        RetryDelayBase = $script:RetryDelayBase
        DockerStartTimeout = $script:DockerStartTimeout
        ContainerHealthTimeout = $script:ContainerHealthTimeout
        BuildOutputMode = $script:BuildOutputMode
    }
}
