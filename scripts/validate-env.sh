#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# Environment Variable Validator
# =============================================================================
# Validates required env vars for given DEPLOY_TARGET and DEPLOY_PROFILE.
# Fails fast with actionable messages naming missing variables.
# Never prints secret values.
# =============================================================================

MISSING=0
WARNINGS=0

# Helper: check if variable is set and non-empty
# Usage: require_var VAR_NAME
require_var() {
    local var_name="$1"
    local var_value="${!var_name:-}"
    
    if [[ -z "$var_value" ]]; then
        echo "❌ Missing required variable: $var_name"
        MISSING=$((MISSING + 1))
    fi
}

# Helper: check if variable is set and non-empty (warn only, don't fail)
# Usage: warn_var VAR_NAME
warn_var() {
    local var_name="$1"
    local var_value="${!var_name:-}"
    
    if [[ -z "$var_value" ]]; then
        echo "⚠️  Warning: missing variable: $var_name"
        WARNINGS=$((WARNINGS + 1))
    fi
}

# Helper: validate variable content
# Usage: validate_var VAR_NAME CONDITION ERROR_MESSAGE
validate_var() {
    local var_name="$1"
    local var_value="${!var_name:-}"
    local condition="$2"
    local error_message="$3"
    
    if [[ -n "$var_value" ]]; then
        if [[ "$condition" == "not_equals" ]]; then
            local forbidden_value="$4"
            if [[ "$var_value" == "$forbidden_value" ]]; then
                echo "❌ Invalid $var_name: $error_message"
                MISSING=$((MISSING + 1))
            fi
        elif [[ "$condition" == "starts_with" ]]; then
            local required_prefix="$4"
            if [[ ! "$var_value" =~ ^"$required_prefix" ]]; then
                echo "❌ Invalid $var_name: $error_message"
                MISSING=$((MISSING + 1))
            fi
        elif [[ "$condition" == "not_in" ]]; then
            local forbidden_values=("${@:4}")
            for forbidden in "${forbidden_values[@]}"; do
                if [[ "$var_value" == "$forbidden" ]]; then
                    echo "❌ Invalid $var_name: $error_message"
                    MISSING=$((MISSING + 1))
                    break
                fi
            done
        fi
    fi
}

echo "========================================="
echo "Environment Variable Validation"
echo "========================================="
echo ""

# Determine deployment context
DEPLOY_TARGET="${DEPLOY_TARGET:-}"
DEPLOY_PROFILE="${DEPLOY_PROFILE:-development}"
SSL_MODE="${SSL_MODE:-}"
ALERT_ENABLED="${ALERT_ENABLED:-false}"

echo "Deploy target: ${DEPLOY_TARGET:-<not set>}"
echo "Deploy profile: $DEPLOY_PROFILE"
echo "SSL mode: ${SSL_MODE:-<not set>}"
echo ""

# =============================================================================
# ALWAYS REQUIRED (all deployments)
# =============================================================================
echo "Checking always-required variables..."
require_var "ENCRYPTION_KEY"
require_var "JWT_SECRET"
require_var "JWT_REFRESH_SECRET"
require_var "POSTGRES_PASSWORD"
require_var "REDIS_PASSWORD"
echo ""

# =============================================================================
# PRODUCTION-SPECIFIC CHECKS
# =============================================================================
if [[ "$DEPLOY_PROFILE" == "production" ]]; then
    echo "Checking production-specific variables..."
    
    # Required in production
    require_var "DOMAIN_NAME"
    require_var "FRONTEND_URL"
    require_var "CORS_ORIGINS"
    
    # Validate DOMAIN_NAME content
    validate_var "DOMAIN_NAME" "not_in" "must not be example.com or localhost in production" "example.com" "localhost"
    
    # Validate CORS_ORIGINS content (must not be wildcard *)
    CORS_ORIGINS_VALUE="${CORS_ORIGINS:-}"
    if [[ -n "$CORS_ORIGINS_VALUE" && "$CORS_ORIGINS_VALUE" == "*" ]]; then
        echo "❌ Invalid CORS_ORIGINS: must not be wildcard (*) in production"
        MISSING=$((MISSING + 1))
    fi
    
    # Validate FRONTEND_URL when SSL_MODE=cloudflare
    if [[ "$SSL_MODE" == "cloudflare" ]]; then
        validate_var "FRONTEND_URL" "starts_with" "must start with https:// when SSL_MODE=cloudflare" "https://"
    fi
    
    echo ""
fi

# =============================================================================
# DEPLOY TARGET: VPS
# =============================================================================
if [[ "$DEPLOY_TARGET" == "vps" ]]; then
    echo "Checking VPS-specific variables..."
    require_var "VPS_HOST"
    require_var "VPS_DEPLOY_PATH"
    echo ""
fi

# =============================================================================
# DEPLOY TARGET: TUNNEL (not VPS)
# =============================================================================
if [[ "$DEPLOY_TARGET" != "vps" && -n "$DEPLOY_TARGET" ]]; then
    echo "Checking tunnel-specific variables..."
    require_var "CLOUDFLARE_TUNNEL_TOKEN"
    echo ""
fi

# =============================================================================
# ALERTING (warn only)
# =============================================================================
if [[ "$ALERT_ENABLED" == "true" ]]; then
    echo "Checking alerting variables (warnings only)..."
    warn_var "ALERT_TELEGRAM_BOT_TOKEN"
    warn_var "ALERT_TELEGRAM_CHAT_ID"
    echo ""
fi

# =============================================================================
# FINAL REPORT
# =============================================================================
echo "========================================="
if [[ $WARNINGS -gt 0 ]]; then
    echo "⚠️  $WARNINGS warning(s)"
fi

if [[ $MISSING -gt 0 ]]; then
    echo "❌ $MISSING variable(s) missing or invalid"
    echo "========================================="
    exit 1
else
    echo "✅ All required variables are set and valid"
    echo "========================================="
    exit 0
fi
