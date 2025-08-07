#!/bin/bash

# Test Coverage Script for assistant-cli
# This script runs comprehensive tests and generates coverage reports

set -e

echo "🚀 Starting comprehensive test suite for assistant-cli..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create reports directory
REPORTS_DIR="test-reports"
mkdir -p $REPORTS_DIR

echo -e "${BLUE}📊 Running tests with coverage...${NC}"

# Run tests with coverage for all packages
go test -v -cover -coverprofile=$REPORTS_DIR/coverage.out ./... | tee $REPORTS_DIR/test-output.txt

# Generate HTML coverage report
echo -e "${BLUE}📈 Generating HTML coverage report...${NC}"
go tool cover -html=$REPORTS_DIR/coverage.out -o $REPORTS_DIR/coverage.html

# Show coverage summary
echo -e "${BLUE}📋 Coverage Summary:${NC}"
go tool cover -func=$REPORTS_DIR/coverage.out | grep -E "(total:|func|github.com)" | tail -20

# Calculate overall coverage percentage
TOTAL_COVERAGE=$(go tool cover -func=$REPORTS_DIR/coverage.out | grep total | awk '{print $3}')
echo -e "${GREEN}🎯 Total Coverage: $TOTAL_COVERAGE${NC}"

# Extract coverage percentage for comparison
COVERAGE_NUM=$(echo $TOTAL_COVERAGE | sed 's/%//')
COVERAGE_THRESHOLD=60

if (( $(echo "$COVERAGE_NUM > $COVERAGE_THRESHOLD" | bc -l) )); then
    echo -e "${GREEN}✅ Coverage above threshold ($COVERAGE_THRESHOLD%)!${NC}"
else
    echo -e "${YELLOW}⚠️  Coverage below threshold ($COVERAGE_THRESHOLD%). Current: $TOTAL_COVERAGE${NC}"
fi

# Generate coverage report by package
echo -e "${BLUE}📦 Coverage by Package:${NC}"
go tool cover -func=$REPORTS_DIR/coverage.out | grep -E "github.com.*:" | sort -k3 -nr

echo -e "${GREEN}✨ Test coverage analysis complete!${NC}"
echo -e "${BLUE}📁 Reports saved to: $REPORTS_DIR/${NC}"
echo -e "${BLUE}🌐 Open $REPORTS_DIR/coverage.html in your browser for detailed coverage visualization${NC}"

# Check for any test failures
if grep -q "FAIL" $REPORTS_DIR/test-output.txt; then
    echo -e "${RED}❌ Some tests failed. Check $REPORTS_DIR/test-output.txt for details.${NC}"
    exit 1
else
    echo -e "${GREEN}✅ All tests passed!${NC}"
fi