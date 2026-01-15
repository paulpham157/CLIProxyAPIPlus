#!/bin/bash

# Simple health check script using curl and jq
# Usage: ./check.sh [URL]

URL="${1:-http://localhost:8080}"
ENDPOINT="${URL}/v0/health"

echo "Checking health at ${ENDPOINT}"
echo ""

RESPONSE=$(curl -s -w "\n%{http_code}" "${ENDPOINT}" 2>&1)
HTTP_CODE=$(echo "${RESPONSE}" | tail -n1)
BODY=$(echo "${RESPONSE}" | head -n-1)

if [ "${HTTP_CODE}" != "200" ]; then
    echo "❌ Health check failed with HTTP ${HTTP_CODE}"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo "✅ Health endpoint responded successfully"
    echo ""
    echo "${BODY}" | python3 -m json.tool 2>/dev/null || echo "${BODY}"
    exit 0
fi

# Parse with jq if available
STATUS=$(echo "${BODY}" | jq -r '.status')
VERSION=$(echo "${BODY}" | jq -r '.version.version')
COMMIT=$(echo "${BODY}" | jq -r '.version.commit' | cut -c1-7)
UPTIME=$(echo "${BODY}" | jq -r '.uptime.human_readable')

TOTAL_PROVIDERS=$(echo "${BODY}" | jq -r '.providers.total')
ACTIVE_PROVIDERS=$(echo "${BODY}" | jq -r '.providers.active')
ERROR_PROVIDERS=$(echo "${BODY}" | jq -r '.providers.error')

TOTAL_REQUESTS=$(echo "${BODY}" | jq -r '.metrics.requests.total')
SUCCESS_RATE=$(echo "${BODY}" | jq -r '.metrics.requests.success_rate')

MEMORY_MB=$(echo "${BODY}" | jq -r '.system.memory_usage_mb')

if [ "${STATUS}" = "healthy" ]; then
    echo "✅ Status: Healthy"
else
    echo "⚠️  Status: ${STATUS}"
fi

echo ""
echo "📊 Service Info:"
echo "   Version: ${VERSION} (${COMMIT})"
echo "   Uptime: ${UPTIME}"
echo ""
echo "🔌 Providers:"
echo "   Total: ${TOTAL_PROVIDERS}"
echo "   Active: ${ACTIVE_PROVIDERS}"
echo "   Errors: ${ERROR_PROVIDERS}"
echo ""
echo "📈 Metrics:"
echo "   Requests: ${TOTAL_REQUESTS}"
echo "   Success Rate: ${SUCCESS_RATE}%"
echo ""
echo "💻 System:"
echo "   Memory: ${MEMORY_MB} MB"

if [ "${STATUS}" != "healthy" ]; then
    exit 1
fi
