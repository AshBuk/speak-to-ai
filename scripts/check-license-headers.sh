#!/bin/bash

# License header validation script
# Checks that all Go files have proper MIT license headers

set -e

echo "🔍 Checking license headers in Go files..."

# Expected patterns
EXPECTED_COPYRIGHT="Copyright (c) 2025 Asher Buk"
EXPECTED_SPDX="SPDX-License-Identifier: MIT"

# Find and check files
missing_files=0
total_files=0

for file in $(find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*"); do
    total_files=$((total_files + 1))
    
    # Check if both copyright and SPDX are present in first 10 lines
    if ! head -10 "$file" | grep -q "$EXPECTED_COPYRIGHT"; then
        echo "❌ Missing copyright: $file"
        missing_files=$((missing_files + 1))
    elif ! head -10 "$file" | grep -q "$EXPECTED_SPDX"; then
        echo "❌ Missing SPDX license: $file"  
        missing_files=$((missing_files + 1))
    fi
done

echo "📊 Processed $total_files Go files"

if [ $missing_files -eq 0 ]; then
    echo "✅ All Go files have valid MIT license headers!"
    exit 0
else
    echo "❌ Found $missing_files files with missing or invalid license headers"
    echo ""
    echo "Expected header format:"
    echo "  // Copyright (c) 2025 Asher Buk"
    echo "  // SPDX-License-Identifier: MIT"
    exit 1
fi