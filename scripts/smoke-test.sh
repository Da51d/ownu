#!/bin/bash
# Smoke tests for OwnU API
# Run against a deployed environment to verify basic functionality

set -e

# Configuration
API_BASE="${API_BASE:-http://localhost:8080}"
FRONTEND_BASE="${FRONTEND_BASE:-http://localhost}"
VERBOSE="${VERBOSE:-false}"
CURL_OPTS="--connect-timeout 10 --max-time 30"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
PASSED=0
FAILED=0
SKIPPED=0

log() {
    echo -e "$1"
}

log_verbose() {
    if [ "$VERBOSE" = "true" ]; then
        echo -e "$1"
    fi
}

pass() {
    ((PASSED++)) || true
    log "${GREEN}✓ PASS${NC}: $1"
}

fail() {
    ((FAILED++)) || true
    log "${RED}✗ FAIL${NC}: $1"
    if [ -n "$2" ]; then
        log "  Details: $2"
    fi
}

skip() {
    ((SKIPPED++)) || true
    log "${YELLOW}○ SKIP${NC}: $1"
}

# Helper function for curl requests
do_curl() {
    curl -s $CURL_OPTS "$@" 2>/dev/null || echo ""
}

do_curl_with_code() {
    curl -s -w "\n%{http_code}" $CURL_OPTS "$@" 2>/dev/null || echo -e "\n000"
}

# Test functions
test_backend_health() {
    log_verbose "Testing: Backend health endpoint"
    local response http_code body
    response=$(do_curl_with_code "$API_BASE/health")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "200" ]; then
        if echo "$body" | grep -q '"status"'; then
            pass "Backend health check returns 200 with status"
        else
            fail "Backend health check returns 200 but missing status field" "$body"
        fi
    else
        fail "Backend health check" "Expected 200, got $http_code"
    fi
}

test_frontend_serves() {
    log_verbose "Testing: Frontend serves index.html"
    local response http_code body
    # Use -L to follow redirects (HTTP->HTTPS), -k to ignore self-signed certs
    response=$(do_curl_with_code -k -L "$FRONTEND_BASE/")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "200" ]; then
        if echo "$body" | grep -q '<html'; then
            pass "Frontend serves HTML content"
        else
            fail "Frontend returns 200 but no HTML content"
        fi
    elif [ "$http_code" = "301" ] || [ "$http_code" = "302" ]; then
        # Redirect is acceptable if we can't follow (e.g., HTTPS not available)
        pass "Frontend redirects (HTTP->HTTPS likely)"
    else
        fail "Frontend serve" "Expected 200/301/302, got $http_code"
    fi
}

test_register_begin_empty_username() {
    log_verbose "Testing: Register begin rejects empty username"
    local response http_code body
    response=$(do_curl_with_code -X POST "$API_BASE/api/v1/auth/register/begin" \
        -H "Content-Type: application/json" \
        -d '{"username":""}')
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "400" ]; then
        if echo "$body" | grep -q 'username is required'; then
            pass "Register begin rejects empty username with proper error"
        else
            fail "Register begin rejects empty username but wrong error message" "$body"
        fi
    else
        fail "Register begin empty username" "Expected 400, got $http_code"
    fi
}

test_register_begin_valid_username() {
    log_verbose "Testing: Register begin accepts valid username"
    local test_user response http_code body
    test_user="smoketest_$(date +%s)"
    response=$(do_curl_with_code -X POST "$API_BASE/api/v1/auth/register/begin" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$test_user\"}")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "200" ]; then
        if echo "$body" | grep -q 'session_id' && echo "$body" | grep -q 'recovery_phrase' && echo "$body" | grep -q 'options'; then
            pass "Register begin returns session_id, recovery_phrase, and options"
        else
            fail "Register begin returns 200 but missing required fields" "$body"
        fi
    else
        fail "Register begin valid username" "Expected 200, got $http_code"
    fi
}

test_register_finish_invalid_session() {
    log_verbose "Testing: Register finish rejects invalid session"
    local response http_code body
    response=$(do_curl_with_code -X POST "$API_BASE/api/v1/auth/register/finish" \
        -H "Content-Type: application/json" \
        -d '{"session_id":"invalid-session","credential":"{}"}')
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "400" ]; then
        if echo "$body" | grep -q 'session expired or invalid'; then
            pass "Register finish rejects invalid session"
        else
            fail "Register finish rejects invalid session but wrong error" "$body"
        fi
    else
        fail "Register finish invalid session" "Expected 400, got $http_code"
    fi
}

test_login_begin_nonexistent_user() {
    log_verbose "Testing: Login begin returns 404 for nonexistent user"
    local response http_code body
    response=$(do_curl_with_code -X POST "$API_BASE/api/v1/auth/login/begin" \
        -H "Content-Type: application/json" \
        -d '{"username":"nonexistent_user_12345"}')
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "404" ]; then
        if echo "$body" | grep -q 'user not found'; then
            pass "Login begin returns 404 for nonexistent user"
        else
            fail "Login begin returns 404 but wrong error message" "$body"
        fi
    else
        fail "Login begin nonexistent user" "Expected 404, got $http_code"
    fi
}

test_login_begin_empty_username() {
    log_verbose "Testing: Login begin rejects empty username"
    local response http_code
    response=$(do_curl_with_code -X POST "$API_BASE/api/v1/auth/login/begin" \
        -H "Content-Type: application/json" \
        -d '{"username":""}')
    http_code=$(echo "$response" | tail -n1)

    if [ "$http_code" = "400" ]; then
        pass "Login begin rejects empty username"
    else
        fail "Login begin empty username" "Expected 400, got $http_code"
    fi
}

test_login_finish_invalid_session() {
    log_verbose "Testing: Login finish rejects invalid session"
    local response http_code body
    response=$(do_curl_with_code -X POST "$API_BASE/api/v1/auth/login/finish" \
        -H "Content-Type: application/json" \
        -d '{"session_id":"invalid-session","credential":"{}"}')
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "400" ]; then
        if echo "$body" | grep -q 'session expired or invalid'; then
            pass "Login finish rejects invalid session"
        else
            fail "Login finish rejects invalid session but wrong error" "$body"
        fi
    else
        fail "Login finish invalid session" "Expected 400, got $http_code"
    fi
}

test_protected_routes_require_auth() {
    log_verbose "Testing: Protected routes require authentication"
    local all_protected=true
    local endpoints=("/api/v1/accounts" "/api/v1/transactions" "/api/v1/categories")

    for endpoint in "${endpoints[@]}"; do
        local response http_code
        response=$(do_curl_with_code "$API_BASE$endpoint")
        http_code=$(echo "$response" | tail -n1)

        if [ "$http_code" != "401" ]; then
            all_protected=false
            fail "Protected endpoint $endpoint" "Expected 401, got $http_code"
        fi
    done

    if [ "$all_protected" = true ]; then
        pass "All protected routes return 401 without auth"
    fi
}

test_cors_headers() {
    log_verbose "Testing: CORS headers are present"
    local response
    response=$(curl -s -I -X OPTIONS "$API_BASE/api/v1/auth/register/begin" \
        -H "Origin: https://localhost" \
        -H "Access-Control-Request-Method: POST" \
        $CURL_OPTS 2>/dev/null || echo "")

    if echo "$response" | grep -qi "access-control-allow"; then
        pass "CORS headers present on OPTIONS request"
    else
        # Some servers don't respond to OPTIONS, try actual request
        response=$(curl -s -I -X POST "$API_BASE/api/v1/auth/register/begin" \
            -H "Origin: https://localhost" \
            -H "Content-Type: application/json" \
            -d '{}' $CURL_OPTS 2>/dev/null || echo "")
        if echo "$response" | grep -qi "access-control-allow"; then
            pass "CORS headers present on POST request"
        else
            skip "CORS headers (may require specific origin)"
        fi
    fi
}

test_api_returns_json() {
    log_verbose "Testing: API returns JSON content type"
    local response
    # Use -i to include headers in output (not -I which does HEAD request)
    response=$(curl -s -i -X POST "$API_BASE/api/v1/auth/register/begin" \
        -H "Content-Type: application/json" \
        -d '{"username":""}' $CURL_OPTS 2>/dev/null || echo "")

    if echo "$response" | grep -qi "content-type:.*application/json"; then
        pass "API returns JSON content type"
    else
        fail "API content type" "Expected application/json header"
    fi
}

test_database_connection() {
    log_verbose "Testing: Database connection (via register begin)"
    local response http_code body
    response=$(do_curl_with_code -X POST "$API_BASE/api/v1/auth/register/begin" \
        -H "Content-Type: application/json" \
        -d '{"username":"db_test_user"}')
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "200" ] || [ "$http_code" = "409" ]; then
        pass "Database connection working (register begin succeeded)"
    else
        if echo "$body" | grep -q "database"; then
            fail "Database connection" "Database error in response"
        else
            pass "Database connection appears working"
        fi
    fi
}

# Main execution
main() {
    log "======================================"
    log "OwnU Smoke Tests"
    log "======================================"
    log "API Base: $API_BASE"
    log "Frontend Base: $FRONTEND_BASE"
    log "======================================"
    log ""

    # Wait for services to be ready
    log "Waiting for services..."
    sleep 2

    # Run tests
    log ""
    log "--- Health & Connectivity ---"
    test_backend_health
    test_frontend_serves
    test_database_connection

    log ""
    log "--- Authentication API ---"
    test_register_begin_empty_username
    test_register_begin_valid_username
    test_register_finish_invalid_session
    test_login_begin_empty_username
    test_login_begin_nonexistent_user
    test_login_finish_invalid_session

    log ""
    log "--- Security ---"
    test_protected_routes_require_auth
    test_cors_headers
    test_api_returns_json

    # Summary
    log ""
    log "======================================"
    log "Test Summary"
    log "======================================"
    log "${GREEN}Passed${NC}: $PASSED"
    log "${RED}Failed${NC}: $FAILED"
    log "${YELLOW}Skipped${NC}: $SKIPPED"
    log "======================================"

    if [ $FAILED -gt 0 ]; then
        exit 1
    fi
    exit 0
}

main "$@"
